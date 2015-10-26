package getstream_test

import (
	"github.com/onedotover/go-getstream"
	a "github.com/stretchr/testify/assert"
	"testing"
)

func TestFeed(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client := ConnectTestClient("")
	feed := client.Feed(TestFeedSlug.Slug, TestFeedSlug.ID)
	activity := TestActivityTarget

	t.Log("adding activity...")
	e := feed.AddActivity(&activity)
	a.NoError(t, e)
	a.NotEmpty(t, activity.ID)

	t.Log("listing added activities...")
	activities := []*getstream.Activity{}
	_, e = feed.Activities(&activities, nil)
	a.NoError(t, e)
	a.NotEmpty(t, activities)
	a.Len(t, activities, 1) // otherwise we might be getting result from another test run.
	a.Equal(t, activity.ID, activities[0].ID)

	t.Log("removing added activity...")
	e = feed.RemoveActivity(activity.ID)
	a.NoError(t, e)

	t.Log("listing added activities again...")
	activities = []*getstream.Activity{}
	_, e = feed.Activities(&activities, nil)
	a.NoError(t, e)
	a.Empty(t, activities)
}

func TestFollow(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client := ConnectTestClient("")

	// Check the follower feed and make sure it's empty
	follower := client.Feed(TestFollowerSlug.Slug, TestFollowerSlug.ID)
	activities := []*getstream.Activity{}
	_, err := follower.Activities(&activities, nil)
	a.NoError(t, err)
	if !a.Len(t, activities, 0) {
		return // The rest of this test depends on just one activity
	}

	// Post an activity to the main test feed
	feed := client.Feed(TestFeedSlug.Slug, TestFeedSlug.ID)
	activity := TestActivitySimple
	err = feed.AddActivity(&activity)
	a.NoError(t, err)

	// Now follow the main feed
	err = follower.Follow(TestFeedSlug.Slug, TestFeedSlug.ID)
	a.NoError(t, err)

	// We should now have content in the follower feed
	activities = []*getstream.Activity{}
	_, err = follower.Activities(&activities, nil)
	a.NoError(t, err)
	if a.Len(t, activities, 1) {
		a.Equal(t, activity.ID, activities[0].ID)
	}

	// Add a second follower so we can test limits and paging
	follower2 := client.Feed(TestFollower2Slug.Slug, TestFollower2Slug.ID)
	follower2.Follow(TestFeedSlug.Slug, TestFeedSlug.ID)

	// The main test feed should have two followers now
	followers, err := feed.Followers(nil)
	a.NoError(t, err)
	if a.Len(t, followers, 2) {
		a.Equal(t, follower2.Slug().String(), followers[0].FeedId)
		a.Equal(t, follower.Slug().String(), followers[1].FeedId)
	}

	// And the follower should be following one feed (the main test feed)
	following, err := follower.Following(nil)
	a.NoError(t, err)
	if a.Len(t, following, 1) {
		a.Equal(t, TestFeedSlug.String(), following[0].TargetId)
	}

	// Test following options
	opt := getstream.FollowingOptions{Limit: 1}
	followers, err = feed.Followers(&opt)
	a.NoError(t, err)
	if a.Len(t, followers, 1) {
		a.Equal(t, follower2.Slug().String(), followers[0].FeedId)
	}

	opt = getstream.FollowingOptions{Limit: 1, Offset: 1}
	followers, err = feed.Followers(&opt)
	a.NoError(t, err)
	if a.Len(t, followers, 1) {
		a.Equal(t, follower.Slug().String(), followers[0].FeedId)
	}

	opt = getstream.FollowingOptions{Filter: TestFeedSlug.String()}
	following, err = follower2.Following(&opt)
	a.NoError(t, err)
	if a.Len(t, following, 1) {
		a.Equal(t, feed.Slug().String(), following[0].TargetId)
	}

	opt = getstream.FollowingOptions{Filter: TestTargetFeedSlug.String()}
	following, err = follower2.Following(&opt)
	a.NoError(t, err)
	a.Len(t, following, 0)

	// Unfollow
	err = follower.Unfollow(TestFeedSlug.Slug, TestFeedSlug.ID)
	a.NoError(t, err)

	err = follower2.Unfollow(TestFeedSlug.Slug, TestFeedSlug.ID)
	a.NoError(t, err)

	// We should be left with a empty feeds
	activities = []*getstream.Activity{}
	_, err = follower.Activities(&activities, nil)
	a.NoError(t, err)
	a.Len(t, activities, 0)

	// And we should be left with no following ties
	followers, err = feed.Followers(nil)
	a.NoError(t, err)
	a.Len(t, followers, 0)

	following, err = follower.Following(nil)
	a.NoError(t, err)
	a.Len(t, following, 0)

	following, err = follower2.Following(nil)
	a.NoError(t, err)
	a.Len(t, following, 0)

	// Clean up the post in the main test feed
	err = feed.RemoveActivity(activity.ID)
	a.NoError(t, err)
}

func TestExtendedActivity(t *testing.T) {
	client := ConnectTestClient("")
	feed := client.Feed(TestFeedSlug.Slug, TestFeedSlug.ID)

	activity := TestActivityExtended
	err := feed.AddActivity(&activity)
	a.NoError(t, err)

	activities := []ExtendedActivity{}
	_, err = feed.Activities(&activities, nil)

	a.NoError(t, err)
	if a.Len(t, activities, 1) {
		a.Equal(t, activity.ID, activities[0].ID)
		a.Equal(t, activity.Title, activities[0].Title)
	}

	feed.RemoveActivity(activity.ID)
}