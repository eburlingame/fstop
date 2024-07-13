package resources

import (
	. "github.com/eburlingame/fstop/models"

	"context"
	"encoding/json"

	"gorm.io/gorm"

	"github.com/maragudk/goqite"
)

type Queue interface {
	AddTask(task ImageImport) error
	Receive() (*goqite.ID, *ImageImport, error)
	Done(id goqite.ID) error
}

type SqliteQueue struct {
	queue *goqite.Queue
}

func InitQueue(gorm_db *gorm.DB) (SqliteQueue, error) {
	db, err := gorm_db.DB()
	if err != nil {
		return SqliteQueue{}, err
	}

	goqite.Setup(context.Background(), db)

	queue := goqite.New(goqite.NewOpts{
		DB:   db,
		Name: "process",
	})

	return SqliteQueue{
		queue,
	}, nil
}

func (q *SqliteQueue) AddTask(task ImageImport) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return q.queue.Send(context.Background(), goqite.Message{
		Body: data,
	})
}

func (q *SqliteQueue) Receive() (*goqite.ID, *ImageImport, error) {
	msg, err := q.queue.Receive(context.Background())
	if err != nil {
		return nil, nil, err
	}

	if msg == nil {
		return nil, nil, nil
	}

	var task ImageImport
	if err := json.Unmarshal(msg.Body, &task); err != nil {
		return nil, nil, err
	}

	return &msg.ID, &task, nil
}

func (q *SqliteQueue) Done(id goqite.ID) error {
	err := q.queue.Delete(context.Background(), id)
	if err != nil {
		return err
	}

	return nil
}
