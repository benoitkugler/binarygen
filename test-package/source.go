package testpackage

// Used to test the special comment below
// binarygen: startOffset=2
type startNoAtSubslice struct{}

// Used to test that aliases are correctly retrieved
type withAlias struct {
	f fl32
}

// Used to test interface support
type withUnion struct {
	version    subtableFlagVersion
	otherField byte
	data       subtableITF `unionField:"version"`
}

type subtableFlagVersion uint16

const (
	subtableFlagVersion1 subtableFlagVersion = iota
	subtableFlagVersion2
)

type subtableITF interface {
	isSubtableITF()
}

type subtableITF1 struct {
	F uint64
}
type subtableITF2 struct {
	F uint8
}

func (subtableITF1) isSubtableITF() {}
func (subtableITF2) isSubtableITF() {}

// Used to test Offset support
type withOffset struct {
	version           uint16
	offsetToSlice     []uint64      `offsetSize:"Offset32" arrayCount:"FirstUint16"`
	offsetToStruct    withFixedSize `offsetSize:"Offset32"`
	a, b, c           byte
	offsetToUnbounded []byte `offsetSize:"Offset16" arrayCount:"ToEnd"`
}

// Used to test []byte support
type WithRawdata struct {
	defaut       []byte // default, external length
	startTo      []byte `subsliceStart:"AtStart"` // external length
	currentToEnd []byte `arrayCount:"ToEnd"`
	startToEnd   []byte `arrayCount:"ToEnd" subsliceStart:"AtStart"`
}

// Used to check that static fields yields
// only one scope
type singleScope struct {
	a, b, c int32
	d       uint32
	e       int64
	g, h    byte
	t       tag
	v       float214 `bin:"optional"`
	w       fl32
	array1  [5]byte
	array2  [5]uint16
}

type multipleScopes struct {
	version uint16
	x, y    int16
	lookups []withFixedSize `arrayCount:"FirstUint16"`
	array2  []uint32        `arrayCount:"FirstUint16"`
}

type customType map[string]int

// customParse is called by the generated parsing code
func (ct *customType) customParse(src []byte) (int, error) {
	return 0, nil
}

type WithOpaque struct {
	f      uint16
	opaque customType `isOpaque:""`
}

type WithSlices struct {
	length uint16
	s1     []varSize `arrayCount:"ComputedField-length"`
}

type varSize struct {
	f1     uint32
	array  []uint32    `arrayCount:"FirstUint16"`
	stucts []withAlias `arrayCount:"FirstUint32"`
}
