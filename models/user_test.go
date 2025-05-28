package models_test

import (
	"flag"
	"log"
	"testing"

	"github.com/intraware/rodan/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var dbURL string

func TestMain(m *testing.M) {
	flag.StringVar(&dbURL, "db", "", "PostgresDB url")
	flag.Parse()
	if dbURL == "" {
		log.Fatal("Do provide the db url for testing as an arg")
	}
	var err error
	db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
		TranslateError:                           true,
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err := db.AutoMigrate(&models.Team{}, &models.User{}); err != nil {
		log.Fatalf("Failed to migrate user and team models: %v", err)
	}
	m.Run()
}

var leaderID int
var teamID int

func TestCreateUser(t *testing.T) {
	t.Log("Creating an user without a team")
	user := models.User{
		Email:          "tester@gmail.com",
		GitHubUsername: "tester",
		Username:       "tester",
		Password:       "testing",
	}
	err := db.Create(&user).Error
	assert.NoError(t, err, "Failed to create an user without a team")
	assert.Greater(t, user.ID, 0, "UserID should be greater than zero")
	leaderID = user.ID
}

func TestCreateTeam(t *testing.T) {
	t.Logf("Creating a team with %d as leader", leaderID)
	team := models.Team{
		Name:     "Tester Team",
		LeaderID: leaderID,
	}
	err := db.Create(&team).Error
	assert.NoError(t, err, "Failed to create with tester as leader")
	teamID = team.ID
	err = db.Model(&models.User{}).Where("id = ?", leaderID).Update("team_id", team.ID).Error
	assert.NoError(t, err, "Failed to add user to team")
	err = db.Preload("Members").First(&team, team.ID).Error
	assert.NoError(t, err, "Failed to load members of the team")
	assert.Len(t, team.Members, 1, "Team should have 1 member")
}

func TestJoinTeam(t *testing.T) {
	t.Logf("Adding members to team %d", teamID)
	another_user := models.User{
		Email:          "anothertester@gmail.com",
		GitHubUsername: "anothertester",
		Username:       "anothertester",
		Password:       "anothertester",
		TeamID:         &teamID,
	}
	var team models.Team
	err := db.Create(&another_user).Error
	assert.NoError(t, err, "Failed to create the user")
	err = db.Preload("Members").First(&team, teamID).Error
	assert.NoError(t, err, "Failed to load members of the team")
	assert.Len(t, team.Members, 2, "Team should have 2 members")
}

func TestDeletedLeader(t *testing.T) {
	t.Logf("Deleting Leader user %d", leaderID)
	var leader_user models.User
	err := db.First(&leader_user, leaderID).Error
	assert.NoError(t, err, "Failed to find the leader user")
	err = db.Delete(&leader_user).Error
	assert.NoError(t, err, "Failed to delete leader user")
	var team models.Team
	err = db.First(&team, leader_user.TeamID).Error
	assert.NoError(t, err, "Failed to get the team")
	leaderID = team.LeaderID
	assert.NotEqual(t, 1, team.LeaderID, "Team Leader ID should not be 1")
}

func TestDeleteTeam(t *testing.T) {
	t.Logf("Deleting team %d", teamID)
	var team models.Team
	err := db.First(&team, teamID).Error
	assert.NoError(t, err, "Failed to get the team")
	err = db.Delete(&team).Error
	assert.NoError(t, err, "Failed to delete the team")
	var user models.User
	err = db.Where("id = ?", leaderID).First(&user).Error
	assert.NoError(t, err, "Failed to get leader user")
	assert.Nil(t, user.TeamID, "TeamID should be null for a deleted team")
}
