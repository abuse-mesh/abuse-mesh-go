package report

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	bolt "github.com/coreos/bbolt"
)

const (
	AbuseReportDBBucket = "abuse-report"
)

//AbuseReportStorage is a interface which defines all methodes that are required in database abstraction layer
type AbuseReportStorage interface {
	GetOneByID([]byte) *AbuseReport
	GetAll() []*AbuseReport
	Insert(*AbuseReport) error
	Update(*AbuseReport) error
	Delete(*AbuseReport) error
}

//AbuseReport is a struct of the Abuse report message
type AbuseReport struct {
	ID                  []byte              //The PGP fingerprint of the message
	SuspectResourceType SuspectResourceType //The suspected resource type
	SuspectResourceID   string              //The suspected resource identifier
	AbuseType           AbuseType           //The type of abuse the resource is accused of
}

type AbuseReportBoltStorage struct {
	DB *bolt.DB
}

func NewAbuseReportBoltStorage(db *bolt.DB) *AbuseReportBoltStorage {

	//Create the abuse report bucket on storage initialization
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(AbuseReportDBBucket))
		return err
	})

	return &AbuseReportBoltStorage{
		DB: db,
	}
}

//GetOneByID returns one abuse report by its id
func (store *AbuseReportBoltStorage) GetOneByID(id []byte) *AbuseReport {
	//Init a empty report
	var report AbuseReport
	var found bool

	//Start a read-only transaction
	store.DB.View(func(tx *bolt.Tx) error {
		//Get the bucket
		bucket := tx.Bucket([]byte(AbuseReportDBBucket))

		//Get the report data for this bucket
		reportData := bucket.Get(id)
		if reportData != nil {
			//Make a new buffer
			var buffer bytes.Buffer
			//Write data to the buffer
			fmt.Fprint(&buffer, reportData)

			//Make a decoder that reads from the buffer
			decoder := gob.NewDecoder(&buffer)

			//Decode the data into the report
			if decoder.Decode(&report) != nil {
				found = true
			}
		}

		return nil
	})

	if found {
		return &report
	} else {
		return nil
	}
}

//GetOneByID returns one abuse report by its id
func (store *AbuseReportBoltStorage) GetAll() []*AbuseReport {

	var reports []*AbuseReport

	//Start a read-only transaction
	store.DB.View(func(tx *bolt.Tx) error {

		//Get the bucket
		bucket := tx.Bucket([]byte(AbuseReportDBBucket))

		bucket.ForEach(func(key []byte, value []byte) error {
			var report AbuseReport

			err := json.Unmarshal(value, &report)

			if err == nil {
				reports = append(reports, &report)
			} else {
				log.Println(err)
			}

			return nil
		})

		return nil
	})

	return reports
}

func (store *AbuseReportBoltStorage) Insert(newReport *AbuseReport) error {

	//Start a transaction
	return store.DB.Update(func(tx *bolt.Tx) error {
		//Get the report bucket
		bucket := tx.Bucket([]byte(AbuseReportDBBucket))

		data, _ := json.Marshal(newReport)

		//Save the encoded struct in the database with the id as key
		bucket.Put(newReport.ID, data)
		return nil
	})

}

func (store *AbuseReportBoltStorage) Update(report *AbuseReport) error {
	return nil
}

func (store *AbuseReportBoltStorage) Delete(report *AbuseReport) error {
	return nil
}
