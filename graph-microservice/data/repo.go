package data

import (
	"context"
	"fmt"
	"log"

	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/ppetar33/twitter-api/graph-microservice/model"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type GraphRepository struct {
	Tracer trace.Tracer
}

const databasePort = "neo4j://graph-db:7687"

func (r GraphRepository) AddEntityInDatabase(ctx context.Context, entity *model.Entity) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.AddEntityInDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))
	err = driver.VerifyConnectivity()
	log.Println(err)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	result, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"CREATE (e:Entity {id: $id, type: $type, name: $name, username: $username}) RETURN e",
			map[string]interface{}{"id": entity.ID, "type": entity.Type, "name": entity.Name, "username": entity.Username})

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		entityRecord, err := result.Single()
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		entityNode, _ := entityRecord.Get("e")
		entity := entityNode.(neo4j.Node)

		return entity.Props, result.Err()
	})

	log.Println(err)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (r GraphRepository) AddFollowRelationInDatabase(ctx context.Context, follow *model.Follow) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.AddFollowRelationInDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	result, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		result, err := transaction.Run(
			"MATCH (a:Entity {id: $requestedBy}), (b:Entity {id: $requestedTo}) CREATE (a)-[r:FOLLOW]->(b) RETURN r",
			map[string]interface{}{"requestedBy": follow.RequestedBy, "requestedTo": follow.RequestedTo})

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		followRecord, err := result.Single()
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		followRel, _ := followRecord.Get("r")
		return followRel, result.Err()
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (r GraphRepository) AddRequestFollowRelationInDatabase(ctx context.Context, follow *model.Follow) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.AddRequestFollowRelationInDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	result, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		result, err := transaction.Run(
			"MATCH (a:Entity {id: $requestedBy}), (b:Entity {id: $requestedTo}) CREATE (a)-[r:FOLLOW_REQUEST]->(b) RETURN r",
			map[string]interface{}{"requestedBy": follow.RequestedBy, "requestedTo": follow.RequestedTo})

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		followRecord, err := result.Single()
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		followRel, _ := followRecord.Get("r")
		return followRel, result.Err()
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (r GraphRepository) RemoveFollowRelationInDatabase(ctx context.Context, unfollow *model.Unfollow) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.RemoveFollowRelationInDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	result, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		result, err := transaction.Run(
			"MATCH (a:Entity {id: $user})-[r:FOLLOW]->(b:Entity {id: $wantToUnfollow}) DELETE r",
			map[string]interface{}{"user": unfollow.User, "wantToUnfollow": unfollow.WantToUnfollow})

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		return "OK", result.Err()
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (r GraphRepository) DeleteFollowRequestRelationInDatabase(ctx context.Context, respond *model.Respond) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.DeleteFollowRequestRelationInDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	result, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		result, err := transaction.Run(
			"MATCH (a:Entity {id: $onRequestFromUser})-[r:FOLLOW_REQUEST]->(b:Entity {id: $user}) DELETE r",
			map[string]interface{}{"user": respond.User, "onRequestFromUser": respond.OnRequestFromUser})

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		return "OK", result.Err()
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (r GraphRepository) AcceptFollowRequestRelationInDatabase(ctx context.Context, respond *model.Respond) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.AcceptFollowRequestRelationInDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	result, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		result, err := transaction.Run(
			"MATCH (a:Entity {id: $onRequestFromUser})-[r:FOLLOW_REQUEST]->(b:Entity {id: $user}) DELETE r",
			map[string]interface{}{"user": respond.User, "onRequestFromUser": respond.OnRequestFromUser})

		_, errFollow := transaction.Run(
			"MATCH (a:Entity {id: $onRequestFromUser}), (b:Entity {id: $user}) CREATE (a)-[r:FOLLOW]->(b) RETURN type(r)",
			map[string]interface{}{"user": respond.User, "onRequestFromUser": respond.OnRequestFromUser})

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		if errFollow != nil {
			span.SetStatus(codes.Error, errFollow.Error())
			return nil, errFollow
		}

		return "OK", result.Err()
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (r GraphRepository) GetAllFollowingFromDatabase(ctx context.Context, user *model.User) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.GetAllFollowingFromDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	var nodes = []*model.Entity{}

	_, err = session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, _ := transaction.Run(
			"MATCH (en {id: $user})-[:FOLLOW*1..1]->(following:Entity) RETURN following AS followings",
			map[string]interface{}{"user": user.ID})

		record, er := result.Collect()
		if er != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, er
		}

		for _, node := range record {

			nd := node.Values[0].(neo4j.Node).Props
			entity := model.Entity{}
			err := mapstructure.Decode(nd, &entity)
			log.Println(err)

			nodes = append(nodes, &entity)
		}

		return nodes, nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return nodes, nil
}

func (r GraphRepository) GetAllFollowersFromDatabase(ctx context.Context, user *model.User) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.GetAllFollowersFromDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	var nodes = []*model.Entity{}

	_, err = session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, _ := transaction.Run(
			"MATCH (en {id: $user})<-[:FOLLOW*1..1]-(following:Entity) RETURN following AS followings",
			map[string]interface{}{"user": user.ID})

		record, er := result.Collect()
		if er != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, er
		}

		for _, node := range record {

			nd := node.Values[0].(neo4j.Node).Props
			entity := model.Entity{}
			err := mapstructure.Decode(nd, &entity)
			log.Println(err)

			nodes = append(nodes, &entity)
		}

		return nodes, nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return nodes, nil
}

func (r GraphRepository) GetAllRecommendedFromDatabase(ctx context.Context, user *model.User) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.GetAllRecommendedFromDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	var nodes = []*model.Entity{}

	_, err = session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, _ := transaction.Run(
			"MATCH (en {id: $user})-[:FOLLOW*2..2]->(recommended:Entity) WHERE NOT recommended.id = $user AND NOT exists((en {id: $user})-[:FOLLOW]->(recommended)) RETURN recommended AS followings",
			map[string]interface{}{"user": user.ID})
		//MATCH (en {id: $user})-[:FOLLOW*2..2]->(recommended:Entity) WHERE NOT recommended.id = $user AND NOT exists((en {id: $user})-[:FOLLOW]->(recommended)) RETURN recommended AS followings
		//MATCH (en {id: $user})-[:FOLLOW*2..2]->(following:Entity) WHERE NOT following.id = $user RETURN following AS followings
		record, er := result.Collect()
		if er != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, er
		}

		for _, node := range record {

			nd := node.Values[0].(neo4j.Node).Props
			entity := model.Entity{}
			_ = mapstructure.Decode(nd, &entity)

			nodes = append(nodes, &entity)
		}

		if len(nodes) == 0 {
			newresult, _ := transaction.Run(
				"MATCH (n:Entity) WHERE NOT n.id = $user AND NOT exists(({id: $user})-[:FOLLOW]->(n)) RETURN n LIMIT 5",
				map[string]interface{}{"user": user.ID})

			newrecord, err := newresult.Collect()
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				return nil, err
			}

			for _, node := range newrecord {

				nd := node.Values[0].(neo4j.Node).Props
				entity := model.Entity{}
				_ = mapstructure.Decode(nd, &entity)

				nodes = append(nodes, &entity)
			}
		}

		return nodes, nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return nodes, nil
}

func (r GraphRepository) GetAllRequestsFromDatabase(ctx context.Context, user *model.User) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.GetAllRequestsFromDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	var nodes = []*model.Entity{}

	_, err = session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, _ := transaction.Run(
			"MATCH (en {id: $user})<-[:FOLLOW_REQUEST*1..1]-(following:Entity) RETURN following AS followings",
			map[string]interface{}{"user": user.ID})

		record, er := result.Collect()
		if er != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, er
		}

		for _, node := range record {

			nd := node.Values[0].(neo4j.Node).Props

			fmt.Println(node.Values[0])
			fmt.Println(nd)

			entity := model.Entity{}
			err := mapstructure.Decode(nd, &entity)
			log.Println(err)

			nodes = append(nodes, &entity)
		}

		return nodes, nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return nodes, nil
}

func (r GraphRepository) DeleteUserFromDatabase(ctx context.Context, user *model.User) (interface{}, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.DeleteUserFromDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, _ := transaction.Run(
			"MATCH (en {id: $user}) DETACH DELETE en",
			map[string]interface{}{"user": user.ID})

		_, er := result.Collect()
		if er != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, er
		}

		return "OK", nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return "OK", nil
}

func (r GraphRepository) CheckRelatinosInDatabase(ctx context.Context, follow *model.Follow) (string, error) {
	ctx, span := r.Tracer.Start(ctx, "GraphRepository.CheckRelatinosInDatabase")

	driver, err := neo4j.NewDriver(databasePort,
		neo4j.BasicAuth("neo4j", "qweqweqwe", ""))

	err = driver.VerifyConnectivity()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	session := driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	var nodes = ""

	_, err = session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, _ := transaction.Run(
			"MATCH (u1 {id: $user1})-[r]->(u2 {id: $user2}) RETURN r",
			map[string]interface{}{"user1": follow.RequestedBy, "user2": follow.RequestedTo})

		record, er := result.Collect()
		if er != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, er
		}

		for _, rel := range record {
			nd := rel.Values[0].(neo4j.Relationship).Type
			nodes = nd
		}

		return nodes, nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return nodes, nil
}
