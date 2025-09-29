package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/internal/models"
	"gorm.io/gorm"
)

func getUserFromContext(userID uint) (models.User, error) {
	if u, ok := shared.UserCache.Get(userID); ok {
		return u, nil
	}
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		return user, err
	}
	shared.UserCache.Set(user.ID, user)
	return user, nil
}

func checkAndUnblockBan(targetID uint, isUser bool, now int64, userRef *models.User) (bool, error) {
	var ban models.BanHistory
	query := models.DB
	if isUser {
		query = query.Where("user_id = ?", targetID)
	} else {
		query = query.Where("team_id = ?", targetID)
	}
	err := query.Order("expires_at DESC").First(&ban).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if err == nil {
		if ban.ExpiresAt > now {
			return true, nil
		}
		go func() {
			if isUser {
				if userRef.Ban {
					userRef.Ban = false
					models.DB.Save(userRef)
					shared.UserCache.Set(userRef.ID, *userRef)
				}
			} else {
				var team models.Team
				var ok bool
				if team, ok = shared.TeamCache.Get(*ban.TeamID); !ok {
					if err := models.DB.First(&team, ban.TeamID).Error; err != nil {
						return
					}
				}
				if team.Ban {
					team.Ban = false
					models.DB.Save(&team)
					shared.TeamCache.Set(team.ID, team)
				}
			}
		}()
	}
	return false, nil
}

func BanMiddleware(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	user, err := getUserFromContext(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		ctx.Abort()
		return
	}
	now := time.Now().Unix()
	if blocked, err := checkAndUnblockBan(user.ID, true, now, &user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user ban"})
		ctx.Abort()
		return
	} else if blocked {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Account is banned"})
		ctx.Abort()
		return
	}
	if user.TeamID != nil {
		if blocked, err := checkAndUnblockBan(*user.TeamID, false, now, &user); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team ban"})
			ctx.Abort()
			return
		} else if blocked {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Team is banned"})
			ctx.Abort()
			return
		}
	}
	ctx.Next()
}
