package pkg3

import (
	"github.com/Djarvur/go-simple-converters/internal/extractor/testdata/pkg1"
	"github.com/Djarvur/go-simple-converters/internal/extractor/testdata/pkg2"
)

var convert_pkg1_pkg2 func(pkg1.Test1Enum) pkg2.Test1Enum

var (
	convert_pkg2_pkg1     func(pkg2.Test1Enum) pkg1.Test1Enum
	convert_pkg2_p_pkg1   func(pkg2.Test1Enum) *pkg1.Test1Enum
	convert_p_pkg1_pkg1   func(*pkg1.Test1Enum) pkg2.Test1Enum
	convert_p_pkg1_p_pkg1 func(*pkg1.Test1Enum) *pkg2.Test1Enum
)
