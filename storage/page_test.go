package btree

import (
	"os"
	"reflect"
	"testing"
)

func TestEncodeDecodeKeyCell(t *testing.T) {

	pg := &page{
		pageID:   10,
		cellType: KeyCell,
		offsets:  []uint16{2, 1, 0, 3},
		freeSize: 3999,
		cells: []interface{}{
			&keyCell{
				key:    123,
				pageID: 3,
			},
			&keyCell{
				key:    12,
				pageID: 8,
			},
			&keyCell{
				key:    1,
				pageID: 6,
			},
			&keyCell{
				key:    1234,
				pageID: 2,
			},
		},
		rightOffset: 1,
		hasLSib:     true,
		hasRSib:     true,
		lSibPageID:  2,
		rSibPageID:  3,
	}

	buf, err := pg.encode()
	if err != nil {
		t.Fatal(err)
	}

	if buf.Len() != pageSize {
		t.Fatalf("page size is not %d bytes, got %d\n", pageSize, buf.Cap())
	}

	actual := &page{}
	if err = actual.decode(buf); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(pg, actual) {
		t.Errorf("Structs are not the same: %v\n%v", pg, actual)
	}
}

func TestEncodeDecodeKeyValueCell(t *testing.T) {

	pg := &page{
		pageID:   10,
		cellType: KeyValueCell,
		offsets:  []uint16{2, 1, 0, 3},
		freeSize: 3949,
		cells: []interface{}{
			&keyValueCell{
				key:        1,
				valueSize:  uint32(len("lorem ipsum")),
				valueBytes: []byte("lorem ipsum"),
			},
			&keyValueCell{
				key:        2,
				valueSize:  uint32(len("dolor sit amet")),
				valueBytes: []byte("dolor sit amet"),
			},
			&keyValueCell{
				key:        3,
				valueSize:  uint32(len("consectetur adipiscing elit")),
				valueBytes: []byte("consectetur adipiscing elit"),
			},
			&keyValueCell{
				key:        4,
				valueSize:  uint32(len("sed do eiusmod")),
				valueBytes: []byte("sed do eiusmod"),
			},
		},
	}

	buf, err := pg.encode()
	if err != nil {
		t.Fatal(err)
	}

	if len(buf.Bytes()) != pageSize {
		t.Fatalf("page size is not %d bytes, got %d\n", pageSize, buf.Cap())
	}

	actual := &page{}
	if err = actual.decode(buf); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(pg, actual) {
		t.Errorf("Structs are not the same: %v\n%v", pg, actual)
	}
}

func TestMemoryStore(t *testing.T) {

	pages := []*page{{}, {}, {}}

	m := &memoryStore{}

	for _, p := range pages {
		if err := m.append(p); err != nil {
			t.Fatal(err)
		}
	}

	for idx, p := range pages {
		fp, err := m.fetch(uint64(idx))
		if err != nil {
			t.Errorf("unable to fetch page at expectedOffset %d", idx)
		}
		if fp != p {
			t.Errorf("page at expectedOffset %d is not the same as the one inserted", idx)
		}
	}
}

func TestFileStore(t *testing.T) {

	fs1 := fileStore{
		path:           "/tmp/page_file",
		nextFreeOffset: pageSize,
	}

	defer os.Remove(fs1.path)

	err := fs1.save()
	if err != nil {
		t.Errorf("unable to save file store: %s", err.Error())
	}

	fs2 := fileStore{
		path: "/tmp/page_file",
	}

	err = fs2.open()
	if err != nil {
		t.Errorf("unable to open file store: %s", err.Error())
	}

	if fs1.nextFreeOffset != fs2.nextFreeOffset {
		t.Errorf("file store branch factors do not match")
	}

	root := &page{
		cellType: KeyValueCell,
	}
	if err := fs2.append(root); err != nil {
		t.Errorf("error appending root: %s", err.Error())
	}
	fs2.setRoot(root)

	root2, err := fs2.getRoot()
	if err != nil {
		t.Errorf("unable to fetch root: %s", err.Error())
	}

	if root.cellType != root2.cellType {
		t.Errorf("root cell types do not match")
	}
}

func TestFindCellByKey(t *testing.T) {
	pages := []*page{
		{
			cellType: KeyCell,
			offsets: []uint16{
				0, 1, 2, 3, 4,
			},
			cells: []interface{}{
				&keyCell{key: 1},
				&keyCell{key: 3},
				&keyCell{key: 5},
				&keyCell{key: 7},
				&keyCell{key: 9},
			},
		},
		{
			cellType: KeyValueCell,
			offsets: []uint16{
				0, 1, 2, 3, 4,
			},
			cells: []interface{}{
				&keyValueCell{key: 1},
				&keyValueCell{key: 3},
				&keyValueCell{key: 5},
				&keyValueCell{key: 7},
				&keyValueCell{key: 9},
			},
		},
	}

	for _, pg := range pages {
		findCellByKeyTestCase(t, pg)
	}
}

func findCellByKeyTestCase(t *testing.T, pg *page) {

	tbl := []struct {
		key            uint32
		expectedOffset int
		expectedFound  bool
	}{
		{
			key:            5,
			expectedOffset: 2,
			expectedFound:  true,
		},
		{
			key:            3,
			expectedOffset: 1,
			expectedFound:  true,
		},
		{
			key:            7,
			expectedOffset: 3,
			expectedFound:  true,
		},
		{
			key:            1,
			expectedOffset: 0,
			expectedFound:  true,
		},
		{
			key:            9,
			expectedOffset: 4,
			expectedFound:  true,
		},
		{
			key:            0,
			expectedOffset: 0,
			expectedFound:  false,
		},
		{
			key:            2,
			expectedOffset: 1,
			expectedFound:  false,
		},
		{
			key:            4,
			expectedOffset: 2,
			expectedFound:  false,
		},
		{
			key:            6,
			expectedOffset: 3,
			expectedFound:  false,
		},
		{
			key:            8,
			expectedOffset: 4,
			expectedFound:  false,
		},
		{
			key:            10,
			expectedOffset: 5,
			expectedFound:  false,
		},
	}

	for _, v := range tbl {
		expectedOffset, expectedFound := pg.findCellOffsetByKey(v.key)
		if expectedOffset != v.expectedOffset || expectedFound != v.expectedFound {
			t.Errorf("[key]: %d [page]: %v [expectedOffset]: %d [actualOffset]: %d [expectedFound]: %t [actualFound]: %t",
				v.key, pg, v.expectedOffset, expectedOffset, v.expectedFound, expectedFound)
		}
	}
}

func TestIsFullKeyValueCellExpectFull(t *testing.T) {
	pg := &page{
		cellType: KeyValueCell,
	}

	for i := 0; i < maxLeafNodeCells; i++ {
		if err := pg.appendCell(uint32(i), []byte("hello")); err != nil {
			t.Fatal(err)
		}
	}

	if !pg.isFull() {
		t.Errorf("leaf node is supposed to be full but is not. max leaf node cells: %d", maxLeafNodeCells)
	}
}

func TestIsFullKeyValueCellExpectNotFull(t *testing.T) {
	pg := &page{
		cellType: KeyValueCell,
	}

	for i := 0; i < maxLeafNodeCells-1; i++ {
		if err := pg.appendCell(uint32(i), []byte("hello")); err != nil {
			t.Fatal(err)
		}
	}

	if pg.isFull() {
		t.Errorf("leaf node is not supposed to be full, but it is. max leaf node cells: %d", maxLeafNodeCells)
	}
}

func TestIsFullKeyCellExpectFull(t *testing.T) {
	pg := &page{
		cellType: KeyCell,
	}

	for i := 0; i < maxInternalNodeCells; i++ {
		if err := pg.appendKeyCell(uint32(i), 1); err != nil {
			t.Fatal(err)
		}
	}

	if !pg.isFull() {
		t.Errorf("internal node is supposed to be full but is not. branch factor: %d", maxInternalNodeCells)
	}
}

func TestIsFullKeyCellExpectNotFull(t *testing.T) {
	pg := &page{
		cellType: KeyCell,
	}

	for i := 0; i < maxInternalNodeCells-1; i++ {
		if err := pg.appendKeyCell(uint32(i), 1); err != nil {
			t.Fatal(err)
		}
	}

	if pg.isFull() {
		t.Errorf("internal node is not supposed to be full, but it is. branch factor: %d", maxInternalNodeCells)
	}
}

func TestSplitKeyValueCell(t *testing.T) {

	pg := &page{
		cellType: KeyValueCell,
	}

	if err := pg.appendCell(0, []byte("hello 0")); err != nil {
		t.Fatal(err)
	}
	if err := pg.appendCell(1, []byte("hello 1")); err != nil {
		t.Fatal(err)
	}
	if err := pg.appendCell(2, []byte("hello 2")); err != nil {
		t.Fatal(err)
	}
	if err := pg.appendCell(3, []byte("hello 3")); err != nil {
		t.Fatal(err)
	}

	newPg := &page{}
	parentKey, err := pg.split(newPg)
	if err != nil {
		t.Fatal(err)
	}

	if parentKey != 2 {
		t.Errorf("parent key is unexpected. actual: %d", parentKey)
	}
	if len(newPg.cells) != 2 {
		t.Errorf("new page is supposed to be half size but is not. size: %d", len(newPg.cells))
	}

	expected := []interface{}{
		&keyValueCell{
			key:        0,
			valueSize:  uint32(len([]byte("hello 0"))),
			valueBytes: []byte("hello 0"),
		},
		&keyValueCell{
			key:        1,
			valueSize:  uint32(len([]byte("hello 1"))),
			valueBytes: []byte("hello 1"),
		},
	}

	for i := 0; i < len(expected); i++ {
		actual := pg.cells[pg.offsets[i]]
		if !reflect.DeepEqual(actual.(*keyValueCell), expected[i].(*keyValueCell)) {
			t.Errorf("key value cell does not match. expected: %+v actual: %+v", expected[i], actual)
		}
	}

	expected = []interface{}{
		&keyValueCell{
			key:        2,
			valueSize:  uint32(len([]byte("hello 2"))),
			valueBytes: []byte("hello 2"),
		},
		&keyValueCell{
			key:        3,
			valueSize:  uint32(len([]byte("hello 3"))),
			valueBytes: []byte("hello 3"),
		},
	}

	for i := 0; i < len(expected); i++ {
		actual := newPg.cells[pg.offsets[i]]
		if !reflect.DeepEqual(actual.(*keyValueCell), expected[i].(*keyValueCell)) {
			t.Errorf("key value cell does not match. expected: %+v actual: %+v", expected[i], actual)
		}
	}
}
