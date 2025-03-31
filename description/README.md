# Оптимизации
### Server
- Заменены `io.ReadAll` на `io.Copy`
```
➜  server git:(iter18) ✗ go tool pprof -top -diff_base=base.pb result.pb
File: main
Type: inuse_space
Time: 2025-03-25 16:56:50 MSK
Duration: 60.01s, Total samples = 4826.40kB 
Showing nodes accounting for 513.99kB, 10.65% of 4826.40kB total
Dropped 9 nodes (cum <= 24.13kB)
      flat  flat%   sum%        cum   cum%
-1026.25kB 21.26% 21.26% -1026.25kB 21.26%  compress/flate.(*huffmanEncoder).generate
  512.05kB 10.61% 10.65%   512.05kB 10.61%  internal/profile.init.func5
  512.01kB 10.61% 0.045%   512.01kB 10.61%  net.IP.String
  512.01kB 10.61% 10.56%   512.01kB 10.61%  net.JoinHostPort (inline)
    4.18kB 0.087% 10.65%     4.18kB 0.087%  compress/flate.(*compressor).initDeflate (inline)
         0     0% 10.65% -1026.25kB 21.26%  compress/flate.(*Writer).Close (inline)
         0     0% 10.65% -1026.25kB 21.26%  compress/flate.(*compressor).close
         0     0% 10.65% -1026.25kB 21.26%  compress/flate.(*compressor).deflate
         0     0% 10.65%     4.18kB 0.087%  compress/flate.(*compressor).init
         0     0% 10.65% -1026.25kB 21.26%  compress/flate.(*compressor).writeBlock
         0     0% 10.65% -1026.25kB 21.26%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 10.65% -1026.25kB 21.26%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0% 10.65%     4.18kB 0.087%  compress/flate.NewWriter (inline)
         0     0% 10.65% -1026.25kB 21.26%  compress/gzip.(*Writer).Close
         0     0% 10.65%     4.18kB 0.087%  compress/gzip.(*Writer).Write
         0     0% 10.65%   512.01kB 10.61%  database/sql.(*DB).QueryContext
         0     0% 10.65%   512.01kB 10.61%  database/sql.(*DB).QueryContext.func1
         0     0% 10.65%   512.01kB 10.61%  database/sql.(*DB).QueryRowContext (inline)
         0     0% 10.65%   512.01kB 10.61%  database/sql.(*DB).conn
         0     0% 10.65%   512.01kB 10.61%  database/sql.(*DB).query
         0     0% 10.65%   512.01kB 10.61%  database/sql.(*DB).retry
         0     0% 10.65%     4.18kB 0.087%  encoding/json.(*Encoder).Encode
         0     0% 10.65% -1026.25kB 21.26%  github.com/gin-contrib/gzip.(*gzipHandler).Handle.func1
         0     0% 10.65%     4.18kB 0.087%  github.com/gin-contrib/gzip.(*gzipWriter).Write
         0     0% 10.65%   512.05kB 10.61%  github.com/gin-contrib/pprof.RouteRegister.WrapH.func10
         0     0% 10.65%   512.01kB 10.61%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 10.65%   512.01kB 10.61%  github.com/jackc/pgx/v5.connect
         0     0% 10.65%   512.01kB 10.61%  github.com/jackc/pgx/v5/pgconn.ConnectConfig
         0     0% 10.65%   512.01kB 10.61%  github.com/jackc/pgx/v5/pgconn.NetworkAddress
         0     0% 10.65%   512.01kB 10.61%  github.com/jackc/pgx/v5/pgconn.buildConnectOneConfigs
         0     0% 10.65%   512.01kB 10.61%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
         0     0% 10.65%   516.18kB 10.70%  github.com/vysogota0399/mem_stats_monitoring/internal/server.(*Server).Start.NewRestUpdateMetricHandler.updateRestMetricHandlerFunc.func4
         0     0% 10.65%   512.01kB 10.61%  github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories.(*Gauge).Create
         0     0% 10.65%   512.01kB 10.61%  github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories.(*Gauge).pushToDB
         0     0% 10.65%   512.01kB 10.61%  github.com/vysogota0399/mem_stats_monitoring/internal/server/service.UpdateMetricService.Call
         0     0% 10.65%   512.01kB 10.61%  github.com/vysogota0399/mem_stats_monitoring/internal/server/service.UpdateMetricService.createGauge
         0     0% 10.65%   512.05kB 10.61%  internal/profile.Parse
         0     0% 10.65%   512.05kB 10.61%  internal/profile.decodeMessage
         0     0% 10.65%   512.05kB 10.61%  internal/profile.parseUncompressed
         0     0% 10.65%   512.05kB 10.61%  internal/profile.unmarshal (inline)
         0     0% 10.65%   512.01kB 10.61%  net.(*TCPAddr).String
         0     0% 10.65%   512.01kB 10.61%  net.ipEmptyString (inline)
         0     0% 10.65%   513.99kB 10.65%  net/http.(*conn).serve
         0     0% 10.65%   512.05kB 10.61%  net/http/pprof.collectProfile
         0     0% 10.65%   512.05kB 10.61%  net/http/pprof.handler.ServeHTTP
         0     0% 10.65%   512.05kB 10.61%  net/http/pprof.handler.serveDeltaProfile
```

### Agent
- По возможнолсти заменен `encoding/json` на `mailru/easyjson`
- Сокращены аллокации объектов, добавлен `sync.Pool`
- Заменены `io.ReadAll` на `io.Copy`

```
➜  agent git:(iter18) ✗ go tool pprof -top -diff_base=base.pb result.pb 
File: main
Type: inuse_space
Time: 2025-03-24 20:06:38 MSK
Showing nodes accounting for 256.04kB, 7.13% of 3592.12kB total
      flat  flat%   sum%        cum   cum%
  768.26kB 21.39% 21.39%   768.26kB 21.39%  go.uber.org/zap/zapcore.newCounters (inline)
 -512.22kB 14.26%  7.13%  -512.22kB 14.26%  runtime.malg
         0     0%  7.13%   768.26kB 21.39%  github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging.MustZapLogger
         0     0%  7.13%   768.26kB 21.39%  go.uber.org/zap.(*Logger).WithOptions
         0     0%  7.13%   768.26kB 21.39%  go.uber.org/zap.Config.Build
         0     0%  7.13%   768.26kB 21.39%  go.uber.org/zap.Config.buildOptions.WrapCore.func5
         0     0%  7.13%   768.26kB 21.39%  go.uber.org/zap.Config.buildOptions.func1
         0     0%  7.13%   768.26kB 21.39%  go.uber.org/zap.New
         0     0%  7.13%   768.26kB 21.39%  go.uber.org/zap.optionFunc.apply
         0     0%  7.13%   768.26kB 21.39%  go.uber.org/zap/zapcore.NewSamplerWithOptions
         0     0%  7.13%   768.26kB 21.39%  main.main
         0     0%  7.13%     -513kB 14.28%  runtime.gopreempt_m (inline)
         0     0%  7.13%     -513kB 14.28%  runtime.goschedImpl
         0     0%  7.13%   768.26kB 21.39%  runtime.main
         0     0%  7.13%      513kB 14.28%  runtime.mcall
         0     0%  7.13%     -513kB 14.28%  runtime.morestack
         0     0%  7.13%  -512.22kB 14.26%  runtime.newproc.func1
         0     0%  7.13%  -512.22kB 14.26%  runtime.newproc1
         0     0%  7.13%     -513kB 14.28%  runtime.newstack
         0     0%  7.13%      513kB 14.28%  runtime.park_m
         0     0%  7.13%  -512.22kB 14.26%  runtime.systemstack
```