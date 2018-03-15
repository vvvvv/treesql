package treesql

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/vilterp/treesql/pkg/lang"
)

type txn struct {
	boltTxn *bolt.Tx
	db      *Database
}

func (s *schema) toScope(txn *txn) (*lang.Scope, *lang.TypeScope) {
	newScope := lang.NewScope(lang.BuiltinsScope)
	newTypeScope := lang.NewTypeScope(lang.BuiltinsTypeScope)
	for _, table := range s.tables {
		if table.isBuiltin {
			continue
		}
		tableRec := table.toVRecord(txn)
		newScope.Add(table.name, tableRec)
		newTypeScope.Add(fmt.Sprintf("%s_t", table.name), tableRec.GetType())
	}
	return newScope, newTypeScope
}

func (table *tableDescriptor) toVRecord(txn *txn) *lang.VRecord {
	attrs := map[string]lang.Value{}

	for _, col := range table.columns {
		if col.name == table.primaryKey {
			// Get iterator.
			iter, err := txn.getTableIterator(table, col.name)
			if err != nil {
				panic(fmt.Sprintf("err getting table iterator: %v", err))
			}
			attrs[col.name] = lang.NewVRecord(map[string]lang.Value{
				"scan": lang.NewVIteratorRef(iter, table.getType()),
				"get":  lang.NewVInt(2), // getter
			})
		}
	}

	return lang.NewVRecord(attrs)
}

// TODO: maybe name BoltIterator
// once there are also virtual table iterators
type tableIterator struct {
	cursor        *bolt.Cursor
	table         *tableDescriptor
	seekedToFirst bool
}

var _ lang.Iterator = &tableIterator{}

func (ti *tableIterator) Next(_ lang.Caller) (lang.Value, error) {
	var key []byte
	var value []byte
	if !ti.seekedToFirst {
		key, value = ti.cursor.First()
		ti.seekedToFirst = true
	} else {
		key, value = ti.cursor.Next()
	}
	if key == nil {
		return nil, lang.EndOfIteration
	}
	// TODO: actually deserialize
	obj, err := ti.table.recordFromBytes(value)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (ti *tableIterator) Close() error {
	// surprisingly, bolt.Cursor doesn't have a .Close()
	return nil
}

func (txn *txn) getTableIterator(table *tableDescriptor, colName string) (*tableIterator, error) {
	colID, err := table.colIDForName(colName)
	if err != nil {
		return nil, err
	}
	tableBucket := txn.boltTxn.Bucket([]byte(table.name))
	if tableBucket == nil {
		return nil, fmt.Errorf("bucket doesn't exist: %s", table.name)
	}
	idxBucket := tableBucket.Bucket(encodeInteger(int32(colID)))
	if idxBucket == nil {
		return nil, fmt.Errorf("bucket doesn't exist: %s/%d", table.name, colID)
	}

	cursor := idxBucket.Cursor()
	//cursor.
	return &tableIterator{
		table:  table,
		cursor: cursor,
	}, nil
}

// TODO: build an vIteratorRef with the right type
// may require using the typ type in the table descriptor
// which would really f*ck things up