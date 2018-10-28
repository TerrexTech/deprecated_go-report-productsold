package report

import (
	"log"
	"strconv"

	"github.com/TerrexTech/go-mongoutils/mongo"
	"github.com/pkg/errors"
)

type ConfigSchema struct {
	Flash     *Flash
	Metric    *Metric
	Inventory *Inventory
}

type DBIConfig struct {
	Hosts               []string
	Username            string
	Password            string
	TimeoutMilliseconds uint32
	Database            string
	Collection          string
}

type DBI interface {
	Collection() *mongo.Collection
	InvAdvSearch(search map[string][]SearchParam) ([]Inventory, error)
}

type DB struct {
	collection *mongo.Collection
}

func GenerateDB(dbConfig DBIConfig, schema interface{}) (*DB, error) {
	config := mongo.ClientConfig{
		Hosts:               dbConfig.Hosts,
		Username:            dbConfig.Username,
		Password:            dbConfig.Password,
		TimeoutMilliseconds: dbConfig.TimeoutMilliseconds,
	}

	client, err := mongo.NewClient(config)
	if err != nil {
		err = errors.Wrap(err, "Error creating DB-client")
		return nil, err
	}

	conn := &mongo.ConnectionConfig{
		Client:  client,
		Timeout: 5000,
	}

	// indexConfigs := []mongo.IndexConfig{
	// 	mongo.IndexConfig{
	// 		ColumnConfig: []mongo.IndexColumnConfig{
	// 			mongo.IndexColumnConfig{
	// 				Name: "item_id",
	// 			},
	// 		},
	// 		IsUnique: true,
	// 		Name:     "item_id_index",
	// 	},
	// 	mongo.IndexConfig{
	// 		ColumnConfig: []mongo.IndexColumnConfig{
	// 			mongo.IndexColumnConfig{
	// 				Name:        "timestamp",
	// 				IsDescOrder: true,
	// 			},
	// 		},
	// 		IsUnique: true,
	// 		Name:     "timestamp_index",
	// 	},
	// }

	// ====> Create New Collection
	collConfig := &mongo.Collection{
		Connection:   conn,
		Database:     dbConfig.Database,
		Name:         dbConfig.Collection,
		SchemaStruct: schema,
		// Indexes:      indexConfigs,
	}
	c, err := mongo.EnsureCollection(collConfig)
	if err != nil {
		err = errors.Wrap(err, "Error creating DB-client")
		return nil, err
	}
	return &DB{
		collection: c,
	}, nil
}

func (d *DB) Collection() *mongo.Collection {
	return d.collection
}

func (db *DB) InvAdvSearch(search map[string][]SearchParam) ([]Inventory, error) {

	var findResults []interface{}
	var err error
	// var type string

	inv := search["inventory"]

	findParams := map[string]interface{}{}

	if inv != nil {
		for _, v := range inv {
			if v.Type == "" {
				err = errors.New("Type required - InvAdvSearch.")
				log.Println(err)
				return nil, err
			}
			if v.Field == "" {
				err = errors.New("Field is required - InvAdvSearch")
				log.Println(err)
				return nil, err
			}

			// if v.Equal == "" && v.Field != "" && v.Type != "" {
			// 	err = errors.New("Equal is required - InvAdvSearch")
			// 	log.Println(err)
			// 	return nil, err
			// }

			if v.Equal == "" && v.LowerLimit == 0 && v.UpperLimit == 0 && v.Field != "" && v.Type != "" {
				err = errors.New("Missing value in equal. No lowerlimit and upperlimit set - InvAdvSearch.")
				log.Println(err)
				return nil, err
			}

			switch v.Type {
			case "string":
				findParams[v.Field] = map[string]string{
					"$eq": v.Equal,
				}

			case "float":
				if v.Equal != "" {
					floatValue, err := strconv.ParseFloat(v.Equal, 64)
					if err != nil {
						err = errors.Wrap(err, "Error converting value of equal to float - InvAdvSearch")
						log.Println(err)
						return nil, err
					}
					findParams[v.Field] = map[string]float64{
						"$eq": floatValue,
					}
				} else {
					limitMap := map[string]map[string]float64{}
					if v.LowerLimit != 0 {
						if limitMap[v.Field] == nil {
							limitMap[v.Field] = map[string]float64{}
						}
						limitMap[v.Field]["$gte"] = v.LowerLimit
					}
					if v.UpperLimit != 0 {
						if limitMap[v.Field] == nil {
							limitMap[v.Field] = map[string]float64{}
						}
						limitMap[v.Field]["$lte"] = v.UpperLimit
					}
					findParams[v.Field] = limitMap[v.Field]
				}

			case "int":
				if v.Equal != "" {
					intValue, err := strconv.ParseInt(v.Equal, 10, 64)
					if err != nil {
						err = errors.Wrap(err, "Error converting equal to int - InvAdvSearch")
						log.Println(err)
						return nil, err
					}
					findParams[v.Field] = map[string]int64{
						"$eq": intValue,
					}
				} else {
					limitMap := map[string]map[string]int64{}
					if v.LowerLimit != 0 {
						if limitMap[v.Field] == nil {
							limitMap[v.Field] = map[string]int64{}
						}
						limitMap[v.Field]["$gte"] = int64(v.LowerLimit)
					}
					if v.UpperLimit != 0 {
						if limitMap[v.Field] == nil {
							limitMap[v.Field] = map[string]int64{}
						}
						limitMap[v.Field]["$lte"] = int64(v.UpperLimit)
					}

					findParams[v.Field] = limitMap[v.Field]
					log.Println(findParams[v.Field])
				}
			}
			if v.Type == "string" {
				findParams[v.Field] = map[string]interface{}{
					"$eq": v.Equal,
				}
			}
		}
	}

	log.Println(findParams, "###123################")

	findResults, err = db.collection.Find(findParams)
	log.Println("FFFFFFFFFFFFFFF")
	for _, r := range findResults {
		log.Printf("%+v", r)
	}

	if err != nil {
		err = errors.Wrap(err, "Error while fetching results from inventory.")
		log.Println(err)
		return nil, err
	}

	//length
	if len(findResults) == 0 {
		msg := "No results found - InvAdvSearch"
		return nil, errors.New(msg)
	}

	inventory := []Inventory{}

	for _, v := range findResults {
		result := v.(*Inventory)
		inventory = append(inventory, *result)
	}
	return inventory, nil
}

// func (db *DB) SearchMetThreshold(threshold float64) (*[]Metric, error) {

// 	findResults, err := db.collection.Find(map[string]interface{}{
// 		"ethylene": map[string]float64{
// 			"$gte": threshold,
// 		},
// 	})

// 	if err != nil {
// 		err = errors.Wrap(err, "Error while searching Metrics db - SearchMetThreshold")
// 		log.Println(err)
// 		return nil, err
// 	}

// 	//length
// 	if len(findResults) == 0 {
// 		msg := "No results found - SearchByDate"
// 		return nil, errors.New(msg)
// 	}

// 	met := []Metric{}

// 	for _, v := range findResults {
// 		result := v.(*Metric)
// 		met = append(met, *result)
// 	}
// 	return &met, nil
// }

// func (db *DB) AddFlashSale(fsale []Flash) ([]*mgo.InsertOneResult, error) {
// 	var insertResult *mgo.InsertOneResult
// 	var getMultipleInserts []*mgo.InsertOneResult

// 	uuid, err := uuuid.NewV4()
// 	if err != nil {
// 		err = errors.Wrap(err, "Unable to generate UUID")
// 		log.Println(err)
// 	}

// 	for _, v := range fsale {
// 		v.FlashID = uuid
// 		v.Timestamp = time.Now().Unix()

// 		if v.FlashID.String() == "00000000-0000-0000-0000-000000000000" {
// 			log.Println("FlashID is empty")
// 			return nil, errors.New("FlashID not found")
// 		}

// 		// log.Println(v.ItemID.String())

// 		// item := v.ItemID.String()

// 		if v.ItemID.String() == "00000000-0000-0000-0000-000000000000" {
// 			log.Println("ItemID is empty")
// 			return nil, errors.New("ItemID not found")
// 		}

// 		if v.UPC == 0 {
// 			log.Println("UPC is empty")
// 			return nil, errors.New("UPC not found")
// 		}

// 		if v.SKU == 0 {
// 			log.Println("SKU is empty")
// 			return nil, errors.New("SKU not found")
// 		}

// 		if v.Name == "" {
// 			log.Println("Name is empty")
// 			return nil, errors.New("Name not found")
// 		}

// 		if v.Origin == "" {
// 			log.Println("Origin is empty")
// 			return nil, errors.New("Origin not found")
// 		}

// 		if v.DeviceID.String() == "00000000-0000-0000-0000-000000000000" {
// 			log.Println("DeviceID is empty")
// 			return nil, errors.New("DeviceID not found")
// 		}

// 		if v.Price == 0 {
// 			log.Println("Price is empty")
// 			return nil, errors.New("Price not found")
// 		}

// 		if v.SalePrice < 0 || v.SalePrice > math.MaxInt64 {
// 			log.Println("Sale Price error. Number is either less than 0 or greater than allowed max value")
// 			return nil, errors.New("Sale Price not found")
// 		}

// 		if v.Ethylene == 0 {
// 			log.Println("Ethylene value is empty")
// 			return nil, errors.New("Ethylene not found")
// 		}

// 		insertResult, err = db.collection.InsertOne(&v)
// 		if err != nil {
// 			err = errors.Wrap(err, "Unable to insert into Flash sale")
// 			log.Println(err)
// 			return nil, err
// 		}

// 		log.Println(insertResult.InsertedID)

// 		getMultipleInserts = append(getMultipleInserts, insertResult)
// 	}

// 	return getMultipleInserts, nil

// }
