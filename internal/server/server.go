package server

import (
    "net/http"
    "time"
    "log"
    "github.com/gin-gonic/gin"
    "github.com/vishal/snow/internal/executor"
    "github.com/vishal/snow/internal/limiter"
)

type Server struct {
    executor *executor.Executor
    limiter  *limiter.RateLimiter
}

type ExecuteRequest struct {
    Language string        `json:"language" binding:"required"`
    Code     string        `json:"code" binding:"required"`
    Timeout  time.Duration `json:"timeout" binding:"required"`
}

type ExecuteResponse struct {
    Output string `json:"output,omitempty"`
    Error  string `json:"error,omitempty"`
}

func New() *Server {
    executor, err := executor.New()
    if err != nil {
        log.Fatalf("failed to create executor: %v", err)
    }

    return &Server{
        executor: executor,
        limiter:  limiter.New(10, 60),
    }
}

func (s *Server) Start(addr string) error {
    r := gin.Default()
    r.POST("/execute", s.HandleExecute)
    return r.Run(addr)
}

// HandleExecute handles code execution requests
func (s *Server) HandleExecute(c *gin.Context) {
    // Get client IP for rate limiting
    clientIP := c.ClientIP()

    // Check rate limit
    if !s.limiter.Allow(clientIP) {
        c.JSON(http.StatusTooManyRequests, ExecuteResponse{
            Error: "rate limit exceeded",
        })
        return
    }

    // Parse request
    var req ExecuteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ExecuteResponse{
            Error: "invalid request: " + err.Error(),
        })
        return
    }

    // Validate timeout
    if req.Timeout < 1*time.Second || req.Timeout > 30*time.Second {
        log.Printf("Received timeout value: %v", req.Timeout);
        c.JSON(http.StatusBadRequest, ExecuteResponse{
            Error: "timeout must be between 1 and 30 seconds",
        })
        return
    }

    // Execute code
    result := s.executor.Execute(c.Request.Context(), executor.ExecutionRequest{
        Language: req.Language,
        Code:     req.Code,
        Timeout:  req.Timeout,
    })


    // Handle execution result
    response := ExecuteResponse{
        Output: result.Output,
    }
    if result.Error != nil {
        response.Error = result.Error.Error()
    }

    // Set appropriate status code
    statusCode := http.StatusOK
    if result.Error != nil {
        statusCode = http.StatusInternalServerError
    }

    c.JSON(statusCode, response)
}