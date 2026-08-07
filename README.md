分析日志的小工具。能从文件读（-f 路径）也能从标准输入接。

三种用法：
  -stats       统计 ERROR/WARN/INFO/DEBUG 各级别出现次数
  -pattern 正则  按正则过滤行，默认保留命中的，加 -keep=false 反向只留没命中的
  -level 级别   只看指定级别的行，比如 -level ERROR
  -top N       统计出现最多的 N 个词
  -json        用 JSON 输出统计结果
  都不带就原样把日志打出来

进阶：
  -context N   命中行（pattern 或 level）前后各多打 N 行上下文，方便看来龙去脉

用法：
  go-logwatch -f app.log -stats
  cat app.log | go-logwatch -pattern ERROR
  cat app.log | go-logwatch -level ERROR -context 2
  cat app.log | go-logwatch -top 10
  cat app.log | go-logwatch -stats -json

测试：
  go test
