分析日志的小工具。能从文件读（-f 路径）也能从标准输入接。

三种用法：
  -stats       统计 ERROR/WARN/INFO/DEBUG 各级别出现次数
  -pattern 正则  按正则过滤行，默认保留命中的，加 -keep=false 反向只留没命中的
  都不带就原样把日志打出来

用法：
  go-logwatch -f app.log -stats
  cat app.log | go-logwatch -pattern ERROR