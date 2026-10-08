package consumer

import (
	"context"
	"example.com/akuanaktehat/pemda-portal/internal/application"
	"example.com/akuanaktehat/pemda-portal/internal/store"
	"github.com/twmb/franz-go/pkg/kgo"
	"testing"
	"time"
)

func TestSQLitePoolExhaustionRetainsOffsetUntilRecovery(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/store.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	processor := &application.Service{Store: db}
	held, err := db.DB.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	r := &kgo.Record{Key: []byte("0199a100-0000-7000-8000-000000000002"), Value: []byte(sample)}
	delivery := &deliveryStub{}
	if err = Process(ctx, r, processor, delivery, 2, 20*time.Millisecond); err == nil {
		t.Fatal("SQLite exhaustion hidden")
	}
	if delivery.dead != 0 || delivery.committed != 0 {
		t.Fatal("offset abandoned during real storage outage")
	}
	held.Close()
	if err = Process(ctx, r, processor, delivery, 2, time.Second); err != nil {
		t.Fatal(err)
	}
	var n int
	if err = db.DB.QueryRow("SELECT count(*) FROM hazard_view").Scan(&n); err != nil || n != 1 || delivery.committed != 1 {
		t.Fatal("recovered record not persisted", err, n)
	}
}
