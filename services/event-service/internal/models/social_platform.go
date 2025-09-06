package models

import (
	"database/sql/driver"
	"errors"
)

type SocialPlatform string

const (
	PlatformFacebook  SocialPlatform = "FACEBOOK"
	PlatformTwitter   SocialPlatform = "TWITTER"
	PlatformInstagram SocialPlatform = "INSTAGRAM"
	PlatformWebsite   SocialPlatform = "WEBSITE"
	PlatformLinkedIn  SocialPlatform = "LINKEDIN"
)

func (sp *SocialPlatform) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*sp = SocialPlatform(v)
	case string:
		*sp = SocialPlatform(v)
	default:
		return errors.New("invalid social platform type")
	}
	return nil
}

func (sp SocialPlatform) Value() (driver.Value, error) {
	return string(sp), nil
}
