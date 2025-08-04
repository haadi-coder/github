# 🐙 Github-CLient 

<div align="center">

*Elegant • Fast • Developer-friendly*

</div>



## ✨ Features

- 🔍 **Search** - Repositories, users, and code with advanced filtering
- 👤 **Users** - Complete profile management and social features  
- 📦 **Repositories** - Full CRUD operations and statistics
- 🐛 **Issues** - Lifecycle management with comments and labels
- 📋 **Pull Requests** - Create, update, merge, and review
- ⚡ **Rate Limiting** - Smart retry with automatic backoff

---

## 📦 Installation

```bash
go get github.com/haadi-coder/github
```

**Requirements:** Go 1.24.4+ and GitHub personal access token

---

## 🚀 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/haadi-coder/github"
)

func main() {
    client, err := github.NewClient(
        github.WithToken("ghp_your_github_token_here"),
        github.WithRateLimitRetry(true),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Get user information
    user, resp, err := client.Users.Get(ctx, "octocat")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("👤 %s (%s) - %d followers\n", user.Name, user.Login, user.Followers)
    fmt.Printf("📊 Rate limit: %d/%d remaining\n", resp.Remaining, resp.Limit)
}
```

---

## 📚 Usage Examples

### 🔍 Search Operations

```go
// Search repositories
repos, _, err := client.Search.Repositories(ctx, "language:go stars:>1000", &github.SearchOptions{
    Sort:  github.String("stars"),
    Order: github.String("desc"),
})

// Search users  
users, _, err := client.Search.Users(ctx, "location:\"San Francisco\"", nil)
```

### 👤 User Management

```go
// Get authenticated user
me, _, err := client.Users.GetAuthenticated(ctx)

// Update profile
user, _, err := client.Users.UpdateAuthenticated(ctx, github.UserUpdateRequest{
    Name: "New Name",
    Bio:  "Updated bio! 🚀",
})

// Follow/unfollow
_, err = client.Users.Follow(ctx, "username")
_, err = client.Users.Unfollow(ctx, "username")

// Get followers
followers, _, err := client.Users.ListAuthenticatedUserFollowers(ctx, nil)
```

### 📦 Repository Operations

```go
// Get repository
repo, _, err := client.Repositories.Get(ctx, "owner", "repo")

// Create repository
newRepo, _, err := client.Repositories.Create(ctx, github.RepositoryCreateRequest{
    Name:        "my-awesome-project",
    Description: "🚀 An awesome project",
    Private:     false,
    AutoInit:    true,
})

// Update repository
updatedRepo, _, err := client.Repositories.Update(ctx, "owner", "repo", github.RepositoryUpdateRequest{
    Description:         "Updated description",
    HasIssues:          true,
    DeleteBranchOnMerge: true,
})

// List contributors
contributors, _, err := client.Repositories.ListContributors(ctx, "owner", "repo", nil)
```

### 🐛 Issue Management

```go
// Get issue
issue, _, err := client.Issues.Get(ctx, "owner", "repo", 1)

// Create issue
newIssue, _, err := client.Issues.Create(ctx, "owner", "repo", &github.IssueCreateRequest{
    Title: "🐛 Bug report",
    Body:  "Something is broken...",
    Labels: []*github.Label{{Name: "bug"}},
})

// Add comment
comment, _, err := client.Issues.CreateComment(ctx, "owner", "repo", 1, github.IssueCommentRequest{
    Body: "Working on this! 🔧",
})

// List issues with filters
issues, _, err := client.Issues.ListByRepo(ctx, "owner", "repo", &github.IssueListOptions{
    State:  github.String("open"),
    Labels: []string{"bug", "priority-high"},
})
```

### 📋 Pull Request Workflow

```go
// Create pull request
pr, _, err := client.PullRequests.Create(ctx, "owner", "repo", &github.PullRequestCreateRequest{
    Title: "✨ Add new feature",
    Head:  "feature-branch",
    Base:  "main",
    Body:  "This PR adds awesome functionality...",
})

// Update pull request
updatedPR, _, err := client.PullRequests.Update(ctx, "owner", "repo", 1, &github.PullRequestUpdateRequest{
    Title: "✨ Updated feature",
    State: "open",
})

// Merge pull request
merge, _, err := client.PullRequests.Merge(ctx, "owner", "repo", 1, &github.MergeRequest{
    CommitTitle:   "✨ Add feature (#1)",
    MergeMethod:   "squash",
})

// List pull requests
prs, _, err := client.PullRequests.List(ctx, "owner", "repo", &github.PullRequestListOptions{
    State: github.String("open"),
    Sort:  github.String("updated"),
})
```

---

## ⚙️ Configuration

### Client Options

```go
client, err := github.NewClient(
    github.WithToken("your-token"),
    github.WithBaseURL("https://api.github.com"),
    github.WithUserAgent("MyApp/1.0"),
    github.WithRateLimitRetry(true),
    github.WithRetryMax(3),
    github.WithHTTPClient(customHTTPClient),
)
```

### Pagination

```go
opts := &github.ListOptions{
    Page:    1,
    PerPage: 50,
}

users, resp, err := client.Users.List(ctx, &github.UsersListOptions{
    ListOptions: opts,
})

// Check pagination info
fmt.Printf("Page %d of %d\n", opts.Page, resp.LastPage)
if resp.NextPage != 0 {
    opts.Page = resp.NextPage
    // Get next page...
}
```

### Error Handling

```go
user, _, err := client.Users.Get(ctx, "nonexistent")
if err != nil {
    if apiErr, ok := err.(*github.APIError); ok {
        switch apiErr.StatusCode {
        case 404:
            fmt.Printf("❌ Not found: %s\n", apiErr.Message)
        case 403:
            fmt.Printf("🚫 Forbidden: %s\n", apiErr.Message)
        case 429:
            fmt.Printf("⏰ Rate limited\n")
        default:
            fmt.Printf("💥 API Error: %d - %s\n", apiErr.StatusCode, apiErr.Message)
        }
    }
}
```

### Rate Limit Monitoring

```go
rateLimits, err := client.RateLimit.Get(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Core API: %d/%d remaining\n", 
    rateLimits.Resources.Core.Remaining, 
    rateLimits.Resources.Core.Limit)
fmt.Printf("Search API: %d/%d remaining\n", 
    rateLimits.Resources.Search.Remaining, 
    rateLimits.Resources.Search.Limit)
```

