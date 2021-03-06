package treesql

import (
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

func (db *Database) validateInsert(insert *Insert) error {
	// does table exist
	tableSpec, ok := db.Schema.Tables[insert.Table]
	if !ok {
		return &NoSuchTable{TableName: insert.Table}
	}
	// can't insert into builtins
	if insert.Table == "__tables__" || insert.Table == "__columns__" {
		return &BuiltinWriteAttempt{TableName: insert.Table}
	}
	// right # fields (TODO: validate types)
	wanted := len(tableSpec.Columns)
	got := len(insert.Values)
	if wanted != got {
		return &InsertWrongNumFields{TableName: insert.Table, Wanted: wanted, Got: got}
	}
	return nil
}

func (conn *Connection) ExecuteInsert(insert *Insert, channel *Channel) error {
	startTime := time.Now()
	table := conn.Database.Schema.Tables[insert.Table]

	// Create record.
	record := table.NewRecord()
	for idx, value := range insert.Values {
		record.SetString(table.Columns[idx].Name, value)
	}
	key := record.GetField(table.PrimaryKey).StringVal

	// Write to table.
	err := conn.Database.BoltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(insert.Table))
		if current := bucket.Get([]byte(key)); current != nil {
			return &RecordAlreadyExists{ColName: table.PrimaryKey, Val: key}
		}
		return bucket.Put([]byte(key), record.ToBytes())
	})
	if err != nil {
		return errors.Wrap(err, "executing insert")
	}

	// Push to live query listeners.
	conn.Database.PushTableEvent(channel, insert.Table, nil, record)
	// Return ack.
	channel.WriteAckMessage("INSERT 1")

	// Record latency.
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	conn.Database.Metrics.insertLatency.Observe(float64(duration.Nanoseconds()))
	// clog.Println(channel, "handled insert in", duration)
	return nil
}
