package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yinheli/database-struct/pkg/model"
	"github.com/yinheli/database-struct/version"
)

var (
	options model.Options
	filters []string

	rootCmd = &cobra.Command{
		Use:   version.AppName,
		Short: version.AppDesc,
		Version: fmt.Sprint(
			"runtime: ", runtime.Version(), ", ",
			"build: ", version.Build, ", ",
			"buildTime: ", version.BuildAt,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.Dsn == "" {
				fmt.Println("Err: missing database dsn")
				os.Exit(1)
			}

			if len(filters) > 0 {
				options.Filters = make([]*model.Filter, 0, len(filters))
				for _, filter := range filters {
					fs := strings.Split(filter, ",")
					var f *model.Filter
					if len(fs) > 1 {
						f = model.NewFilter(fs[0], fs[1])
					} else if len(fs) == 1 {
						f = model.NewFilter("", fs[0])
					}

					if f != nil {
						options.Filters = append(options.Filters, f)
					}
				}
			}

			if options.Verbose {
				b, _ := json.MarshalIndent(&options, "", "  ")
				fmt.Println("using options:\n", string(b))
			}

			tables, err := model.DbStruct(&options)
			if err != nil {
				return err
			}
			return model.Generate(&options, tables)
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&options.DbType, "dbType", "d", model.DbTypeMySQL, "database type: "+strings.Join([]string{model.DbTypeMySQL, model.DbTypePostgreSQL}, ","))
	rootCmd.Flags().StringVarP(&options.Dsn, "dsn", "c", "", "database dsn, e.g. root:123456@(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local")
	rootCmd.Flags().BoolVarP(&options.GenGormTag, "gorm", "", true, "generate gorm tags for model")
	rootCmd.Flags().BoolVarP(&options.GormV1, "gormv1", "", false, "set gorm v1 for model, default v2")
	rootCmd.Flags().BoolVarP(&options.GenJsonTag, "json", "", true, "generate json tags for model")
	rootCmd.Flags().StringVarP(&options.HtmlFile, "html", "", "", "generate html report file")
	rootCmd.Flags().StringVarP(&options.ModelDir, "dir", "", "", "generate go model files to dir")
	rootCmd.Flags().StringVarP(&options.ModelPackageName, "pkg", "", "model", "go model package name")
	rootCmd.Flags().BoolVarP(&options.ModelSingleFile, "single", "", true, "generate go model code all in one file, use `--single=false` to turnoff")
	rootCmd.Flags().StringSliceVarP(&filters, "filter", "f", nil, "filter table with table prefix and pattern, e.g: app_,app_%")
	rootCmd.Flags().StringSliceVarP(&options.Exclude, "exclude", "e", nil, "exclude table name, not support pattern yet")
	rootCmd.Flags().BoolVarP(&options.Verbose, "verbose", "", false, "enable verbose, show more log message")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		println(err)
		os.Exit(1)
	}
}
