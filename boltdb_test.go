package storage

import (
	"testing"

	"time"

	"github.com/boltdb/bolt"
	"github.com/mailhog/data"
)

func TestCreateBoltDB(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "db in tmp",
			args: args{
				path: "/tmp/db",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateBoltDB(tt.args.path); got == nil {
				t.Errorf("CreateBoltDB() = %v", got)
			} else {
				got.db.Close()
			}
		})
	}
}

func TestBoltDB_Store(t *testing.T) {
	db := CreateBoltDB("/tmp/db")
	defer db.db.Close()

	type fields struct {
		db     *bolt.DB
		bucket []byte
	}
	type args struct {
		m *data.Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "good test",
			fields: fields{
				db:     db.db,
				bucket: db.bucket,
			},
			args: args{
				m: &data.Message{
					ID: "",
					From: &data.Path{
						Relays:  nil,
						Mailbox: "from@test.com",
						Domain:  "",
						Params:  "",
					},
					To: []*data.Path{
						{
							Relays:  nil,
							Mailbox: "to@test.com",
							Domain:  "",
							Params:  "",
						},
					},
					Content: &data.Content{
						Headers: nil,
						Body:    "test body",
						Size:    9,
						MIME:    &data.MIMEBody{},
					},
					Created: time.Now(),
					MIME:    &data.MIMEBody{},
					Raw: &data.SMTPMessage{
						From: "from@test.com",
						To:   []string{"to@test.com"},
						Data: "test body",
						Helo: "",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BoltDB{
				db:     tt.fields.db,
				bucket: tt.fields.bucket,
			}
			got, err := b.Store(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("BoltDB.Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("BoltDB.Store() = %v", got)
			}
		})
	}
}

func TestBoltDB_Count(t *testing.T) {
	b := CreateBoltDB("/tmp/db")
	defer b.db.Close()

	tests := []struct {
		name string
	}{
		{
			name: "good test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.Count(); got == 0 {
				t.Errorf("BoltDB.Count() = %v", got)
			}
		})
	}
}

func TestBoltDB_Search(t *testing.T) {
	b := CreateBoltDB("/tmp/db")
	defer b.db.Close()

	type args struct {
		kind  string
		query string
		start int
		limit int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good test",
			args: args{
				kind:  "to",
				query: "to@",
				start: 0,
				limit: 100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := b.Search(tt.args.kind, tt.args.query, tt.args.start, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("BoltDB.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil || len(*got) == 0 {
				t.Errorf("BoltDB.Search() got = %v", got)
			}
			if got1 == 0 {
				t.Errorf("BoltDB.Search() got1 = %v", got1)
			}
		})
	}
}

func TestBoltDB_List(t *testing.T) {
	b := CreateBoltDB("/tmp/db")
	defer b.db.Close()

	type args struct {
		start int
		limit int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good test",
			args: args{
				start: 0,
				limit: 100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.List(tt.args.start, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("BoltDB.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil || len(*got) == 0 {
				t.Errorf("BoltDB.List() = %v", got)
			}
		})
	}
}

func TestBoltDB_Load(t *testing.T) {
	b := CreateBoltDB("/tmp/db")
	defer b.db.Close()

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good test",
			args: args{
				id: "1",
			},
			wantErr: false,
		},
		{
			name: "not found test",
			args: args{
				id: "112312",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.Load(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("BoltDB.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("BoltDB.Load() = %v", got)
			}
		})
	}
}

func TestBoltDB_DeleteOne(t *testing.T) {
	b := CreateBoltDB("/tmp/db")
	defer b.db.Close()

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good test",
			args: args{
				id: "1",
			},
			wantErr: false,
		},
		{
			name: "good test #2",
			args: args{
				id: "1234",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := b.DeleteOne(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("BoltDB.DeleteOne() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoltDB_DeleteAll(t *testing.T) {
	b := CreateBoltDB("/tmp/db")
	defer b.db.Close()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "good test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := b.DeleteAll(); (err != nil) != tt.wantErr {
				t.Errorf("BoltDB.DeleteAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
