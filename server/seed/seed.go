package seed

import (
	"fmt"
	"mig/db"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SeederPostgreSQL struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	faker   *gofakeit.Faker
}

func NewSeederPostgreSQL(pool *pgxpool.Pool, queries *db.Queries) (*SeederPostgreSQL, error) {
	if pool == nil {
		return nil, fmt.Errorf("missing pool")
	}

	if queries == nil {
		return nil, fmt.Errorf("missing queries")
	}

	seeder := &SeederPostgreSQL{
		pool:    pool,
		queries: queries,
		faker:   gofakeit.New(0),
	}

	return seeder, nil
}

var UsersUUIDs = [22]string{
	"ed46e813-ac7b-43a9-854e-fa3f842527e3",
	"132054bf-8795-4b64-b5cb-9646abbbd034",
	"c6b09b3e-de3f-4e80-afc0-a0bf55427535",
	"6fb8b632-76c8-4562-8a55-14b11a4916f9",
	"adcbf943-82a3-4994-9a63-af3aaeee4bad",
	"b761ecbe-7921-4451-be61-9faec23f6855",
	"bcced743-ffc4-4c9e-8261-67cdd8ae7891",
	"9def80ec-9185-42a0-b2f2-8a06660d9a60",
	"d6f26e98-82f2-404f-99a0-9203415a0d53",
	"9025e7f7-2876-4760-8e3c-425078f66c94",
	"f487e7c4-27ad-4ca1-89cd-bc6d89cf6328",
	"89d62302-481b-4562-a9e8-4f5847cafa30",
	"418b8705-27e6-48dc-8568-116da28f5604",
	"bd3f8966-734e-4504-a1bf-78f1c7ad257f",
	"c6ef97dd-ee85-407a-9034-5dde1328e91d",
	"ffce5770-878a-49f0-8e7e-edfa2adbcaac",
	"a850ba39-1a2d-4dca-9f02-ca6353535c46",
	"721e4edd-f15f-4b08-9ad2-9d5058b40c5f",
	"5909d614-eef7-4480-8757-34eff7691ca2",
	"12e5af72-cc3e-4b29-96d3-1f45592898dc",
	"7b0aa607-9e2d-4fcf-8687-8d7f123479e5",
	"986104e4-2fed-48d9-9fa0-01e827a1a234",
}

var ChatroomUUIDs = [22]string{
	"930d5ea1-334b-4323-9be9-3849ed93f289",
	"3777469c-438d-46e1-a3e8-0bd54b16b6c2",
	"65a4f2bc-2cf0-4a9e-b155-33528dda001e",
	"623a5787-eba7-46c8-9d1b-46c015b48951",
	"985c542c-e0ec-4d64-978e-e8303fb9e96f",
	"ab985d07-55ad-451b-b083-faa3674afa56",
	"68cb545e-a64f-415b-a009-e53b14996957",
	"ae72258a-f3c3-4151-946f-d55b028087d8",
	"07ed9498-4999-47cb-9ee7-f903b4645978",
	"f73b530e-85b0-445e-adab-65d66abaafb8",
	"1ba0817c-329b-4d8c-b673-444c51adbe96",
	"bf7d58bf-b92c-4342-b482-5f41b8336ec1",
	"6facbdcd-8b93-47a6-a94f-e905e420ac29",
	"a3e6162a-e1bb-41ce-8581-51733f2ccf86",
	"0ba6f2f3-768f-42b1-b3e4-e17d3c03bd66",
	"0601059e-eac7-492d-b0dd-f1fb543ed035",
	"6f29558a-d257-4967-90b9-57a9c296558b",
	"1e1df990-4b61-42a5-b0eb-f5266081e930",
	"a9f362ec-da37-4d49-8812-093b81f13df7",
	"8d230b91-1f90-4230-bec7-36db552c49c9",
	"3905a29d-8025-4b50-bc84-170f40d9ccab",
	"b52a198c-b2dd-46b1-b04c-0f6eee2d9248",
}
