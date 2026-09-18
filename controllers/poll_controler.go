package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pulse-backend/services"
)

type PollController struct {
	pollService *services.PollService
	voteService *services.VoteService
}

func NewPollController(
	pollService *services.PollService,
	voteService *services.VoteService,
) *PollController {
	return &PollController{
		pollService: pollService,
		voteService: voteService,
	}
}

type CreatePollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

func (p *PollController) CreatePoll(c *gin.Context) {
	var request CreatePollRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	userIDValue, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDValue.(string)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user identity",
		})
		return
	}

	poll, err := p.pollService.CreatePoll(
		userID,
		services.CreatePollInput{
			Question: request.Question,
			Options:  request.Options,
		},
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "poll created successfully",
		"poll":        poll,
		"id":          poll.ID,
		"code":        poll.Code,
		"question":    poll.Question,
		"options":     poll.Options,
		"votes":       poll.Votes,
		"status":      poll.Status,
		"totalVotes":  poll.TotalVotes,
		"percentages": poll.Percentages,
	})
}

func (p *PollController) GetPoll(c *gin.Context) {
	pollID := c.Param("id")

	poll, err := p.pollService.GetPollDTO(c.Request.Context(), pollID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "poll not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"poll":        poll,
		"id":          poll.ID,
		"code":        poll.Code,
		"question":    poll.Question,
		"options":     poll.Options,
		"votes":       poll.Votes,
		"status":      poll.Status,
		"totalVotes":  poll.TotalVotes,
		"percentages": poll.Percentages,
	})
}

func (p *PollController) GetMyPolls(c *gin.Context) {
	userIDValue, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDValue.(string)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user identity",
		})
		return
	}

	polls, err := p.pollService.GetMyPollDTOs(c.Request.Context(), userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch polls",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"polls": polls,
	})
}

type VoteRequest struct {
	OptionID    string `json:"optionId"`
	OptionIndex *int   `json:"optionIndex"`
	VoterID     string `json:"voterId"`
}

func (p *PollController) Vote(c *gin.Context) {
	pollID := c.Param("id")

	var request VoteRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	poll, err := p.voteService.Vote(
		pollID,
		request.OptionID,
		request.OptionIndex,
		request.VoterID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "vote submitted successfully",
		"poll":        poll,
		"id":          poll.ID,
		"code":        poll.Code,
		"question":    poll.Question,
		"options":     poll.Options,
		"votes":       poll.Votes,
		"status":      poll.Status,
		"totalVotes":  poll.TotalVotes,
		"percentages": poll.Percentages,
	})
}

func (p *PollController) ClosePoll(c *gin.Context) {
	pollID := c.Param("id")

	userIDValue, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDValue.(string)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user identity",
		})
		return
	}

	poll, err := p.pollService.ClosePoll(userID, pollID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "poll closed successfully",
		"poll":    poll,
	})
}

func (p *PollController) GetResults(c *gin.Context) {
	pollID := c.Param("id")

	poll, err := p.pollService.GetPollDTO(c.Request.Context(), pollID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "poll not found",
		})
		return
	}

	results, totalVotes, _ := p.voteService.GetResults(pollID)

	c.JSON(http.StatusOK, gin.H{
		"poll":        poll,
		"id":          poll.ID,
		"code":        poll.Code,
		"question":    poll.Question,
		"options":     poll.Options,
		"votes":       poll.Votes,
		"totalVotes":  totalVotes,
		"percentages": poll.Percentages,
		"results":     results,
	})
}