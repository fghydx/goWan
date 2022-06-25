# goWan
golang网络框架

tcp包 -> tcpReadPacket包 -> tcpCustomHead 包三个包可以分开用，一个在一个基础上加功能
tcp包实现最基本的接收发送
tcpReadPacket包在tcp基础上实现包头，包体分开读
tcpCustomHead在tcpReadPacket包基础上实现了一个自定义的包头(ID+包体长度+包体)，并可以注册路由，达到不同请求由不同函数执行的目的。
# Bug还有不少，慢慢修复