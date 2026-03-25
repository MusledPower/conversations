package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

	dsn := "postgres://postgres:12345678@localhost:8003/conv?sslmode=disable"

	ctx := context.Background()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = seed(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("seed completed")
}

func seed(ctx context.Context, db *pgxpool.Pool) error {

	adminID := uuid.Must(uuid.NewV7())
	user1ID := uuid.Must(uuid.NewV7())
	user2ID := uuid.Must(uuid.NewV7())

	room1ID := uuid.Must(uuid.NewV7())
	room2ID := uuid.Must(uuid.NewV7())

	schedule1ID := uuid.Must(uuid.NewV7())
	schedule2ID := uuid.Must(uuid.NewV7())

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx, `
	insert into users (id,email,role,created_at)
	values
	($1,'admin@test.com','admin', now()),
	($2,'user1@test.com','user',now()),
	($3,'user2@test.com','user',now())
	on conflict do nothing
	`,
		adminID,
		user1ID,
		user2ID,
	)

	if err != nil {
		return err
	}

	// ROOMS

	_, err = tx.Exec(
		ctx, `
	insert into rooms (id,name,description,capacity,created_at)
	values
	($1,'Переговорная A','Большая комната',12,now()),
	($2,'Переговорная B','Малая комната',6,now())
	on conflict do nothing
	`,
		room1ID,
		room2ID,
	)

	if err != nil {
		return err
	}

	// SCHEDULES

	_, err = tx.Exec(
		ctx, `
	insert into schedules (id,room_id,days_of_week,start_time,end_time)
	values
	($1,$2,'{1,2,3,4,5}','09:00','18:00'),
	($3,$4,'{6,7}','10:00','16:00')
	on conflict do nothing
	`,
		schedule1ID,
		room1ID,
		schedule2ID,
		room2ID,
	)

	if err != nil {
		return err
	}

	for i := 0; i < 7; i++ {

		day := time.Now().AddDate(0, 0, i)

		for h := 9; h < 18; h++ {

			start := time.Date(
				day.Year(),
				day.Month(),
				day.Day(),
				h,
				0,
				0,
				0,
				time.UTC,
			)

			end := start.Add(time.Hour)

			slotID := uuid.Must(uuid.NewV7())

			_, err = tx.Exec(
				ctx, `
			insert into slots (id,room_id,start_time,end_time)
			values ($1,$2,$3,$4)
			on conflict do nothing
			`,
				slotID,
				room1ID,
				start,
				end,
			)

			if err != nil {
				return err
			}
		}
	}

	slotRows, err := tx.Query(
		ctx,
		`
		select id
		from slots
		where start_time > now()
		limit 2
		`,
	)

	if err != nil {
		return err
	}

	var slotIDs []uuid.UUID

	for slotRows.Next() {

		var id uuid.UUID

		err = slotRows.Scan(&id)
		if err != nil {
			return err
		}

		slotIDs = append(slotIDs, id)
	}

	for _, slotID := range slotIDs {

		bookingID := uuid.Must(uuid.NewV7())

		_, err = tx.Exec(
			ctx, `
		insert into bookings
		(id,slot_id,user_id,status,conference_link,created_at)
		values
		($1,$2,$3,'active','https://meet.test/link',now())
		on conflict do nothing
		`,
			bookingID,
			slotID,
			user1ID,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
