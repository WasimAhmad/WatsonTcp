using BenchmarkDotNet.Running;
using System;

namespace WatsonTcp.Benchmark
{
    public class Program
    {
        public static void Main(string[] args)
        {
            Console.WriteLine("Starting WatsonTcp Benchmarks...");
            // This will run all benchmark classes in the assembly.
            // For specific benchmarks: BenchmarkRunner.Run<MessageThroughputBenchmark>();
            BenchmarkSwitcher.FromAssembly(typeof(Program).Assembly).Run(args);
            Console.WriteLine("WatsonTcp Benchmarks complete.");
        }
    }
}
