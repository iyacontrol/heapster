
| metric名称  |      说明 |
|----------|:-------------:|
|cpu/limit |	CPU hard limit，单位为毫秒|
|cpu/usage |	全部Core的CPU累计使用时间|
|cpu/usage_rate|	全部Core的CPU累计使用率，单位为毫秒|
|filesystem/limit	|文件系统总空间限制，单位为字节|
|filesystem/usage|	文件系统已用的空间，单位为字节|
|memory/limit	|Memory hard limit，单位为字节|
|memory/major_page_faults	|major page faults数量|
|memory/major_page_faults_rate	|每秒的major page faults数量|
|memory/node_allocatable	|Node可分配的内存容量|
|memory/node_capacity	|Node的内存容量|
|memory/node_reservation	|Node保留的内存share|
|memory/node_utilization	|Node的内存使用值|
|memory/page_faults	page |faults数量|
|memory/page_faults_rate	|每秒的page faults数量|
|memory/request	|Memory request，单位为字节|
|memory/usage	|总内存使用量|
|memory/working_set	|总的Working set |usage，Working set是指不会被kernel移除的内存
|network/rx	|累计接收的网络流量字节数|
|network/rx_errors	|累计接收的网络流量错误数|
|network/rx_errors_rate	|每秒接收的网络流量错误数|
|network/rx_rate	|每秒接收的网络流量字节数|
|network/tx	|累计发送的网络流量字节数|
|network/tx_errors	|累计发送的网络流量错误数|
|network/tx_errors_rate	|每秒发送的网络流量错误数|
|network/tx_rate	|每秒发送的网络流量字节数|
|uptime	|容器启动总时长|