package options

var DbList = []string{"urtyg_ai_agent"}                             // 数据库列表
var ModelRootPath = "./internal/fmc-go-agent-server/model/dalm/dbm" // 数据库模型根路径
var DbIfRootPath = "./internal/fmc-go-agent-server/dal/db/dbif"     // 数据库接口根路径
var DbGdbRootPath = "./internal/fmc-go-agent-server/dal/db/gdb"     // 数据库接口Gdb实现根路径
var DbUdbRootPath = "./internal/fmc-go-agent-server/dal/db/udb"     // 数据库接口Gdb实现根路径
var ModelDirName = "model"                                          // 数据库迁移文件根路径
var AutoMigrateFileName = "auto_migrate.gen.go"                     // 自动迁移文件名
var GenIfFileName = "genif.gen.go"                                  // 生成通用接口文件名
var IfFileName = "if.go"                                            // 生成基础与自定义接口文件名
var IfGenFileName = "if.gen.go"                                     // 生成基础与自定义接口文件名
var GenIFImpFileName = "gen.gen.go"                                 // 生成通用接口实现文件名
var SelfIFImpFileName = "self.go"                                   // gdb和udb生成通用接口实现文件名
var DbGenFileName = "db.gen.go"                                     // gdb和udb生成通用接口实现文件名
var GdbImpFileName = "gdb.gen.go"                                   // gdb生成通用接口实现文件名
var UdbImpFileName = "udb.gen.go"                                   // udb生成通用接口实现文件名
