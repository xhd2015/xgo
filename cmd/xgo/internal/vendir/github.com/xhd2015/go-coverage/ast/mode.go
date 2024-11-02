package ast

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/golang.org/x/tools/go/packages"



type LoadMode string

var (
	LoadMode_NeedName  = LoadMode(packages.NeedName.String())
	LoadMode_NeedFiles = LoadMode(packages.NeedFiles.String())

	LoadMode_NeedCompiledGoFiles = LoadMode(packages.NeedCompiledGoFiles.String())

	LoadMode_NeedImports = LoadMode(packages.NeedImports.String())

	LoadMode_NeedDeps = LoadMode(packages.NeedDeps.String())

	LoadMode_NeedExportFile = LoadMode(packages.NeedExportFile.String())

	LoadMode_NeedTypes = LoadMode(packages.NeedTypes.String())

	LoadMode_NeedSyntax = LoadMode(packages.NeedSyntax.String())

	LoadMode_NeedTypesInfo = LoadMode(packages.NeedTypesInfo.String())

	LoadMode_NeedTypesSizes = LoadMode(packages.NeedTypesSizes.String())

	LoadMode_NeedModule = LoadMode(packages.NeedModule.String())

	LoadMode_NeedEmbedFiles = LoadMode(packages.NeedEmbedFiles.String())

	LoadMode_NeedEmbedPatterns = LoadMode(packages.NeedEmbedPatterns.String())
)

var ModeMap = map[LoadMode]packages.LoadMode{
	LoadMode_NeedName:  packages.NeedName,
	LoadMode_NeedFiles: packages.NeedFiles,
	LoadMode_NeedCompiledGoFiles: packages.NeedCompiledGoFiles,
	LoadMode_NeedImports: packages.NeedImports,
	LoadMode_NeedDeps: packages.NeedDeps,
	LoadMode_NeedExportFile: packages.NeedExportFile,
	LoadMode_NeedTypes: packages.NeedTypes,
	LoadMode_NeedSyntax: packages.NeedSyntax,
	LoadMode_NeedTypesInfo: packages.NeedTypesInfo,
	LoadMode_NeedTypesSizes: packages.NeedTypesSizes,
	LoadMode_NeedModule: packages.NeedModule,
	LoadMode_NeedEmbedFiles: packages.NeedEmbedFiles,
	LoadMode_NeedEmbedPatterns: packages.NeedEmbedPatterns,
}