package receiverifgen

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var Handler = &cli.Command{
	Name:  "genif",
	Usage: "根据结构体的接收器方法，生成对应的If文档（需要手动补充包引用）",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "path",
			Usage:    "指定结构体所在的文件夹路径",
			Aliases:  []string{},
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		err := generateInterface(c.String("path"))
		if err != nil {
			return err
		}
		return nil
	},
}

// generateInterface 从接收器方法生成接口定义
func generateInterface(dirPath string) error {
	// 创建文件集和方法集合
	fset := token.NewFileSet()
	methods := make(map[string][]*ast.FuncDecl)

	// 遍历目录中的所有 .go 文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			// 解析文件
			node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				fmt.Printf("Warning: failed to parse file %s: %v\n", path, err)
				return nil
			}

			// 提取方法集
			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					if x.Recv != nil && len(x.Recv.List) > 0 {
						recvType := getTypeName(x.Recv.List[0].Type)
						if recvType != "" {
							methods[recvType] = append(methods[recvType], x)
						}
					}
				}
				return true
			})
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	// 生成接口定义
	var interfaceCode strings.Builder
	for typeName, methodList := range methods {
		interfaceName := typeName + "If"
		interfaceCode.WriteString(fmt.Sprintf("type %s interface {\n", interfaceName))
		for _, method := range methodList {
			methodSig := getMethodSignature(method)
			interfaceCode.WriteString(fmt.Sprintf("\t%s\n", methodSig))
		}
		interfaceCode.WriteString("}\n\n")
	}

	// 写入文件
	outputPath := filepath.Join(dirPath, "if.go")
	err = os.WriteFile(outputPath, []byte(interfaceCode.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write interface file: %w", err)
	}

	fmt.Printf("Generated interfaces saved to: %s\n", outputPath)
	return nil
}

// getTypeName 获取接收器类型的名称
func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// getMethodSignature 获取方法签名
func getMethodSignature(method *ast.FuncDecl) string {
	sig := method.Name.Name + "("
	if method.Type.Params != nil {
		for i, param := range method.Type.Params.List {
			if i > 0 {
				sig += ", "
			}
			sig += paramNamesAndTypes(param)
		}
	}
	sig += ")"
	if method.Type.Results != nil {
		sig += " ("
		for i, result := range method.Type.Results.List {
			if i > 0 {
				sig += ", "
			}
			sig += paramNamesAndTypes(result)
		}
		sig += ")"
	}
	return sig
}

// paramNamesAndTypes 获取参数或返回值的名称和类型
func paramNamesAndTypes(field *ast.Field) string {
	names := ""
	if len(field.Names) > 0 {
		for i, name := range field.Names {
			if i > 0 {
				names += ", "
			}
			names += name.Name
		}
		names += " "
	}
	return names + typeToString(field.Type)
}

// typeToString 将类型转换为字符串
func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

