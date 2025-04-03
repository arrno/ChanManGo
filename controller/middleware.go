package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/rs/zerolog/log"
)

// Auth middleware validates tokens and sets context values
func (c *Controller) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header or query parameter
		tokenString := extractToken(r)

		token := domain.NewToken("", "")
		ok := token.Deserialize(tokenString)
		if !ok {
			RespondWithError(w, http.StatusUnauthorized, "Authentication token is missing or invalid", nil)
			return
		}

		// Validate token and get client permissions
		clientGroup, permission, isAdmin, err := c.db.GetClientAuthByToken(*token)
		if err != nil || clientGroup == nil {
			log.Error().Err(err).Str("token", "<scrub>").Msg("Invalid authentication token")
			RespondWithError(w, http.StatusUnauthorized, "Invalid authentication token", nil)
			return
		}

		// Set headers for downstream handlers
		r.Header.Add("portalId", domain.MAIN_PORTAL)
		r.Header.Add("groupId", clientGroup.GroupId)
		// Need group name
		r.Header.Add("clientId", clientGroup.ClientId)

		// Add admin flag if applicable
		if isAdmin {
			r.Header.Add("admin", "true")
		}

		// Add read/write permissions to headers
		permission.AddToHeaders(r)

		// Log access for audit purposes
		log.Info().
			Str("clientId", clientGroup.ClientId).
			Str("groupId", clientGroup.GroupId).
			Bool("admin", isAdmin).
			Bool("read", permission.Read).
			Bool("write", permission.Write).
			Str("path", r.URL.Path).
			Msg("Authenticated request")

		// Continue with the next handler
		next.ServeHTTP(w, r)
	}
}

// Helper function to extract token from various sources
func extractToken(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Split on space to handle "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 {
			return parts[1]
		}
		return authHeader
	}

	// Check query parameters
	token := r.URL.Query().Get("Authorization")
	if token != "" {
		return token
	}

	// Check form values for POST requests
	return r.FormValue("token")
}

// channelContext middleware validates channel access and existence
func (c *Controller) channelContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		channelName := r.URL.Query().Get("channel")
		topics := r.URL.Query().Get("topics")
		// Set headers for downstream handlers
		r.Header.Add("channelName", channelName)
		r.Header.Add("topics", topics)

		// Extract context values set by auth middleware
		groupId := r.Header.Get("groupId")
		portalId := r.Header.Get("portalId")
		clientId := r.Header.Get("clientId")

		// Validate channel name
		if channelName == "" {
			RespondWithError(w, http.StatusBadRequest, "Channel name is required", nil)
			return
		}

		// Generate full channel ID
		channelId := domain.CompositeId(channelName, groupId)

		// Fetch channel from database
		channel, err := c.db.GetChannel(channelId)
		if err != nil {
			if errors.Is(err, datastore.ErrNotFound) {
				RespondWithError(w, http.StatusNotFound, "Channel not found", nil)
			} else {
				log.Error().Err(err).Str("channelId", channelId).Msg("Error fetching channel")
				RespondWithError(w, http.StatusInternalServerError, "Error fetching channel", err)
			}
			return
		}

		// Verify channel belongs to the expected group and portal
		if channel.GroupId != groupId || channel.PortalId != portalId {
			log.Warn().
				Str("clientId", clientId).
				Str("channelId", channelId).
				Str("expectedGroupId", groupId).
				Str("actualGroupId", channel.GroupId).
				Msg("Channel group/portal mismatch")
			RespondWithError(w, http.StatusForbidden, "You don't have access to this channel", nil)
			return
		}

		// Add channel object to request context
		r.Header.Set("channelId", channel.Id)

		// Continue with next handler
		next.ServeHTTP(w, r)
	}
}

// rateLimit middleware prevents abuse by limiting request frequency
func (c *Controller) rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client identifier - prefer authenticated client ID if available
		clientId := r.Header.Get("clientId")
		identifier := clientId

		// Fall back to IP address if no client ID
		if identifier == "" {
			identifier = r.RemoteAddr
			// Handle potential X-Forwarded-For or similar headers in production
			forwardedFor := r.Header.Get("X-Forwarded-For")
			if forwardedFor != "" {
				// Use the leftmost IP which is the original client
				parts := strings.Split(forwardedFor, ",")
				identifier = strings.TrimSpace(parts[0])
			}
		}

		// Check if this client is allowed to make another request
		if !c.rateLimiter.Allow(identifier) {
			log.Warn().
				Str("identifier", identifier).
				Str("path", r.URL.Path).
				Msg("Rate limit exceeded")

			// Return 429 Too Many Requests with proper headers
			w.Header().Set("Retry-After", "60") // Suggest retry after 1 minute
			RespondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.", nil)
			return
		}

		// Process the request
		next.ServeHTTP(w, r)
	})
}

// admin middleware ensures only admin users can access protected routes
func admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin := r.Header.Get("admin")
		if admin != "true" {
			log.Warn().
				Str("clientId", r.Header.Get("clientId")).
				Str("path", r.URL.Path).
				Msg("Unauthorized admin access attempt")

			RespondWithError(w, http.StatusForbidden, "This action requires administrator privileges", nil)
			return
		}

		// Continue with admin-protected handler
		next.ServeHTTP(w, r)
	}
}

// methodGuard ensures endpoints only respond to the specified HTTP method
func methodGuard(allowedMethod string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			// Set the Allow header to indicate which methods are permitted
			w.Header().Set("Allow", allowedMethod)
			RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
			return
		}

		next(w, r)
	}
}
