using BenchmarkDotNet.Attributes;
using BenchmarkDotNet.Jobs;
using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
using WatsonTcp;

namespace WatsonTcp.Benchmark
{
    [SimpleJob(RuntimeMoniker.Net60)] // Example: Targeting .NET 6. Adjust if needed.
    [MemoryDiagnoser] // To get memory allocation stats
    public class MessageThroughputBenchmark
    {
        private WatsonTcpServer _server;
        private WatsonTcpClient _client;
        private TaskCompletionSource<bool> _tcs;
        private int _messagesToReceive;
        private byte[] _messagePayload;

        private const string ServerIp = "127.0.0.1";
        private const int ServerPort = 9000;

        [Params(128, 1024, 16384)] // 128B, 1KB, 16KB
        public int MessageSize { get; set; }

        [Params(1000, 10000)]
        public int NumberOfMessages { get; set; }

        [GlobalSetup]
        public void GlobalSetup()
        {
            Console.WriteLine($"GlobalSetup: MessageSize={MessageSize}, NumberOfMessages={NumberOfMessages}");

            _messagePayload = new byte[MessageSize];
            new Random(42).NextBytes(_messagePayload); // Fill with deterministic data

            _server = new WatsonTcpServer(ServerIp, ServerPort);
            _server.Events.MessageReceived += Server_MessageReceived;
            _server.Events.ClientConnected += (s, e) => Console.WriteLine($"Benchmark: Client {e.Client.Guid} connected.");
            _server.Events.ClientDisconnected += (s, e) => Console.WriteLine($"Benchmark: Client {e.Client.Guid} disconnected.");
            _server.Start();
            Console.WriteLine("Benchmark: Server started.");

            _client = new WatsonTcpClient(ServerIp, ServerPort);
            _client.Events.ServerConnected += (s, e) => Console.WriteLine("Benchmark: Client connected to server.");
            _client.Events.ServerDisconnected += (s, e) => Console.WriteLine("Benchmark: Client disconnected from server.");
            // _client.Settings.Logger = (sev, msg) => Console.WriteLine($"[Client-{sev}] {msg}");
            // _server.Settings.Logger = (sev, msg) => Console.WriteLine($"[Server-{sev}] {msg}");
            _client.Connect();
            Console.WriteLine("Benchmark: Client connection attempt finished.");
        }

        private void Server_MessageReceived(object sender, MessageReceivedEventArgs e)
        {
            // Basic check
            if (e.Data.Length != MessageSize)
            {
                 Console.WriteLine($"Error: Server received message of size {e.Data.Length}, expected {MessageSize}");
                 // How to fail the benchmark run from here? For now, log and continue counting.
            }

            if (Interlocked.Decrement(ref _messagesToReceive) == 0)
            {
                _tcs?.TrySetResult(true);
            }
        }

        [IterationSetup]
        public void IterationSetup()
        {
            Console.WriteLine($"IterationSetup: Preparing for {NumberOfMessages} messages.");
            _messagesToReceive = NumberOfMessages;
            _tcs = new TaskCompletionSource<bool>(TaskCreationOptions.RunContinuationsAsynchronously);

            // Ensure client is still connected, especially if previous iteration had issues
            if (!_client.Connected)
            {
                Console.WriteLine("IterationSetup: Client was disconnected, attempting reconnect.");
                _client.Connect(); // This might need more robust handling or fail the iteration
                if (!_client.Connected)
                {
                    throw new InvalidOperationException("Client could not connect in IterationSetup.");
                }
            }
        }

        [Benchmark]
        public async Task SendMessages_ClientToServer()
        {
            for (int i = 0; i < NumberOfMessages; i++)
            {
                // Not awaiting each send to simulate rapid fire, relying on TCP and library's internal send queue.
                // If SendAsync becomes a bottleneck or throws, this might need adjustment.
                _ = _client.SendAsync(_messagePayload);
            }

            // Wait for the server to confirm receipt of all messages
            try
            {
                // It's possible the TCS is set very quickly if messages are processed fast.
                // Add a timeout to prevent hangs if something goes wrong.
                var timeoutTask = Task.Delay(TimeSpan.FromSeconds(30)); // 30 second timeout
                var completedTask = await Task.WhenAny(_tcs.Task, timeoutTask);

                if (completedTask == timeoutTask || !_tcs.Task.IsCompletedSuccessfully)
                {
                    Console.WriteLine($"Timeout or error waiting for messages. Messages left: {_messagesToReceive}");
                    // This will likely make the benchmark result very high (bad), indicating an issue.
                    // Or throw an exception:
                    // throw new TimeoutException($"Benchmark timed out waiting for {NumberOfMessages} messages. {_messagesToReceive} remaining.");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Exception during wait: {ex}");
                // Rethrow or handle to ensure benchmark failure is noted
                throw;
            }
        }

        [GlobalCleanup]
        public void GlobalCleanup()
        {
            Console.WriteLine("GlobalCleanup: Stopping client and server.");
            _client?.Disconnect();
            _client?.Dispose();
            _server?.Stop();
            _server?.Dispose();
            Console.WriteLine("GlobalCleanup: Client and server stopped and disposed.");
        }
    }
}
