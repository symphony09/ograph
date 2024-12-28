# OGraph 性能测试报告

## OGraph VS CGraph [ 20241228 updated ]

### 测试说明

为了更好的衡量 OGraph 的性能，本次测试以同类高性能 C++ 调度框架 CGraph 为基准。

在相同条件下运行 4 个场景并对比性能差异。

#### 测试代码

为保证公平性，OGraph 测试用例采用与 CGraph 相同的设计。包含4个场景：

1.  32 个节点，节点之间没有任何依赖关系

2.  32 个节点，节点之间串行连接（n1->n2->...->n32）

3.  6 个节点，构成简单的 DAG 图

4.  64 个节点，分为 8 层，每层 8个节点，层与层全连接

OGraph 脚本：[github.com/symphony09/ograph/test/benchmark.sh](https://github.com/symphony09/ograph/blob/main/test/benchmark.sh)

OGraph 测试用例：[github.com/symphony09/ograph/test/benchmark\_test.go](https://github.com/symphony09/ograph/blob/main/test/benchmark_test.go)

CGraph 脚本：[github.com/ChunelFeng/CGraph/CGraph-run-performance-tests.sh](https://github.com/ChunelFeng/CGraph/blob/main/CGraph-run-performance-tests.sh)

CGraph 测试用例：[github.com/ChunelFeng/CGraph/test/Performance](https://github.com/ChunelFeng/CGraph/tree/main/test/Performance)

### 测试结果

#### OGraph

    [pipeline 循环执行 100 万次]
    goos: linux
    goarch: amd64
    pkg: github.com/symphony09/ograph/test
    cpu: AMD Ryzen 5 5600G with Radeon Graphics         
    BenchmarkConcurrent_32-8         1000000              4308 ns/op            1592 B/op         16 allocs/op
    BenchmarkSerial_32-8             1000000               281.7 ns/op           120 B/op          4 allocs/op
    BenchmarkComplex_6-8             1000000              2762 ns/op            1048 B/op         21 allocs/op
    BenchmarkConnect_8x8-8           1000000              8333 ns/op            2553 B/op         16 allocs/op
    PASS
    ok      github.com/symphony09/ograph/test       15.692s

#### CGraph

与OGraph 每个用例执行 100 万次不同，CGraph 4个用例分别执行了 50万、100 万、100 万、10 万次 。

    **** Running test-performance-01 ****
    [2024-12-23 21:07:45.992] [test_performance_01]: time counter is : [4102.16] ms
     
    **** Running test-performance-02 ****
    [2024-12-23 21:07:46.567] [test_performance_02]: time counter is : [572.51] ms
     
    **** Running test-performance-03 ****
    [2024-12-23 21:07:50.611] [test_performance_03]: time counter is : [4042.16] ms
     
    **** Running test-performance-04 ****
    [2024-12-23 21:07:51.958] [test_performance_04]: time counter is : [1345.02] ms
     
     congratulations, automatic run CGraph performance test finish...

### 结果分析

CGraph 4 个用例的 pipeline 执行次数分为 50w, 100w, 100w, 10w。

换算出每次执行时间分别为 8204 ns/op, 572 ns/op, 4042 ns/op, 13450 ns/op。

|                | CGraph（基准）  | OGraph（本项目）        |
| :------------- | :---------- | :----------------- |
| 场景一（无连接32节点）   | 8204 ns/op  | 4308 ns/op（+90.4%） |
| 场景二（串行连接32节点）  | 572 ns/op   | 281.7 ns/op（+103%） |
| 场景三（简单DAG 6节点） | 4042 ns/op  | 2762 ns/op（+46.3%） |
| 场景四（8x8全连接）    | 13450 ns/op | 8333 ns/op（+61.4%） |

从 4 个场景的测试用例结果来看，OGraph 与 CGraph 性能在同一量级，甚至在不少场景还能领先。

同时，在测试过程中，OGraph 的 CPU 占用率也明显低于 CGraph。

### 并发性能

OGraph pipeline 支持并发执行，在相同场景下，并发测试结果如下：

|                | OGraph（pipeline 循环执行） | OGraph（pipeline 并行执行） |
| :------------- | :-------------------- | :-------------------- |
| 场景一（无连接32节点）   | 4308 ns/op            | 3138 ns/op（+37.2%）    |
| 场景二（串行连接32节点）  | 281.7 ns/op           | 57.25 ns/op（+392%）    |
| 场景三（简单DAG 6节点） | 2762 ns/op            | 747.8 ns/op（+269.3%）  |
| 场景四（8x8全连接）    | 8333 ns/op            | 5136 ns/op（+62.2%）    |

测试脚本： [ograph/test/benchmark\_parallel.sh at main · symphony09/ograph](https://github.com/symphony09/ograph/blob/main/test/benchmark_parallel.sh)

### 附录

#### 测试硬件

【CPU】AMD 5600G - 6 核心 12 线程

【内存】DDR4 32 GB

#### 测试软件

【系统】Linux 6.11 - Fedora Workstation 41

【OGraph】v0.6.1 - Go 1.23.4

【CGraph】v2.6.2 - GCC14.2.1
