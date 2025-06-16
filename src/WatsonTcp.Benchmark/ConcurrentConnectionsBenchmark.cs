using BenchmarkDotNet.Attributes;
using BenchmarkDotNet.Jobs;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using WatsonTcp;

namespace WatsonTcp.Benchmark
{
    [SimpleJob(RuntimeMoniker.Net60)]
    [MemoryDiagnoser]
    public class ConcurrentConnectionsBenchmark
    {
        private WatsonTcpServer _server = null!;
        private List<WatsonTcpClient> _clients = null!;
        private TaskCompletionSource<bool> _allMessagesSentTcs = null!; // Signaled when all clients have sent all messages
        private TaskCompletionSource<bool> _allMessagesReceivedTcs = null!; // Signaled when server received all messages

        private int _totalMessagesToReceive;
        private volatile int _serverMessagesReceivedCount;
        private volatile int _clientsReadyCount;

        private const string ServerIp = "127.0.0.1";
        private const int ServerPort = 9001; // Different port to avoid conflict if run in parallel

        [Params(10, 50)] // Number of concurrent clients
        public int NumberOfClients { get; set; }

        [Params(10, 100)] // Messages each client will send
        public int MessagesPerClient { get; set; }

        public int MessageSize { get; set; } = 128; // Fixed message size for this benchmark for simplicity

        private byte[] _messagePayload = null!;

        [GlobalSetup]
        public void GlobalSetup()
        {
            Console.WriteLine($"GlobalSetup: NumberOfClients={NumberOfClients}, MessagesPerClient={MessagesPerClient}");
            _messagePayload = new byte[MessageSize];
            new Random(42).NextBytes(_messagePayload);

            _server = new WatsonTcpServer(ServerIp, ServerPort);
            _server.Events.ClientConnected += Server_ClientConnected;
            _server.Events.MessageReceived += Server_MessageReceived;
            _server.Start();
            Console.WriteLine("Benchmark (Concurrent): Server started.");
            _clients = new List<WatsonTcpClient>();
        }

        private void Server_ClientConnected(object? sender, ConnectionEventArgs e)
        {
            Interlocked.Increment(ref _clientsReadyCount);
            // Console.WriteLine($"Benchmark (Concurrent): Client connected. Total ready: {_clientsReadyCount}");
        }

        private void Server_MessageReceived(object? sender, MessageReceivedEventArgs e)
        {
            if (Interlocked.Increment(ref _serverMessagesReceivedCount) == _totalMessagesToReceive)
            {
                _allMessagesReceivedTcs?.TrySetResult(true);
            }
        }

        [IterationSetup]
        public async Task IterationSetup()
        {
            Console.WriteLine($"IterationSetup (Concurrent): Preparing for {NumberOfClients} clients, {MessagesPerClient} msg/client.");
            _clients.Clear(); // Ensure clean list for each iteration
            _serverMessagesReceivedCount = 0;
            _clientsReadyCount = 0;
            _totalMessagesToReceive = NumberOfClients * MessagesPerClient;
            _allMessagesReceivedTcs = new TaskCompletionSource<bool>(TaskCreationOptions.RunContinuationsAsynchronously);
            _allMessagesSentTcs = new TaskCompletionSource<bool>(TaskCreationOptions.RunContinuationsAsynchronously);

            // Connect clients
            List<Task> connectTasks = new List<Task>();
            for (int i = 0; i < NumberOfClients; i++)
            {
                var client = new WatsonTcpClient(ServerIp, ServerPort);
                // client.Settings.Logger = (sev, msg) => Console.WriteLine($"[Client{i}-{sev}] {msg}");
                _clients.Add(client);
                connectTasks.Add(Task.Run(() => client.Connect())); // Connect in parallel
            }
            await Task.WhenAll(connectTasks); // Wait for all Connect() calls to complete initiation

            // Wait for all clients to be registered by the server (simple way to check server readiness)
            // This relies on ClientConnected events incrementing _clientsReadyCount
            int checks = 0;
            while (_clientsReadyCount < NumberOfClients && checks < 200) // Max 20 seconds wait
            {
                await Task.Delay(100);
                checks++;
            }
            if (_clientsReadyCount < NumberOfClients)
            {
                throw new InvalidOperationException($"Timeout: Only {_clientsReadyCount}/{NumberOfClients} clients connected to server in IterationSetup.");
            }
            Console.WriteLine($"IterationSetup (Concurrent): All {NumberOfClients} clients connected and acknowledged by server.");
        }

        [Benchmark]
        public async Task ConcurrentClientSends()
        {
            List<Task> clientSendTasks = new List<Task>();

            foreach (var client in _clients)
            {
                clientSendTasks.Add(Task.Run(async () =>
                {
                    for (int i = 0; i < MessagesPerClient; i++)
                    {
                        await client.SendAsync(_messagePayload).ConfigureAwait(false);
                    }
                }));
            }

            await Task.WhenAll(clientSendTasks); // Wait for all clients to finish sending
            _allMessagesSentTcs.TrySetResult(true); // Signal that all sends are done

            // Now wait for server to receive all messages
            var timeoutTask = Task.Delay(TimeSpan.FromSeconds(60)); // 60 second timeout
            var completedTask = await Task.WhenAny(_allMessagesReceivedTcs.Task, timeoutTask);

            if (completedTask == timeoutTask || !_allMessagesReceivedTcs.Task.IsCompletedSuccessfully)
            {
                 Console.WriteLine($"Timeout or error waiting for all messages on server. Received: {_serverMessagesReceivedCount}/{_totalMessagesToReceive}");
                // throw new TimeoutException($"Benchmark timed out waiting for server to receive all messages. Received: {_serverMessagesReceivedCount}/{_totalMessagesToReceive}");
            }
        }

        [IterationCleanup]
        public async Task IterationCleanup()
        {
            Console.WriteLine("IterationCleanup (Concurrent): Disconnecting clients.");
            List<Task> disconnectTasks = new List<Task>();
            foreach (var client in _clients)
            {
                disconnectTasks.Add(Task.Run(() =>
                {
                    client.Disconnect();
                    client.Dispose();
                }));
            }
            await Task.WhenAll(disconnectTasks); // Ensure all clients are disconnected and disposed
            _clients.Clear();
            Console.WriteLine("IterationCleanup (Concurrent): Clients disconnected.");
        }

        [GlobalCleanup]
        public void GlobalCleanup()
        {
            Console.WriteLine("GlobalCleanup (Concurrent): Stopping server.");
            _server?.Stop();
            _server?.Dispose();
            Console.WriteLine("GlobalCleanup (Concurrent): Server stopped and disposed.");
        }
    }
}
