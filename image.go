package wumpkit

import (
	"slices"
	"strconv"
)

// Base URLs for Discord CDN and media assets.
const (
	ImageBaseURL = "https://cdn.discordapp.com/"
	MediaBaseURL = "https://media.discordapp.net/"
)

// ImageSize defines supported image sizes for Discord assets.
type ImageSize int

const (
	ImageSizeDefault ImageSize = 0
	ImageSize16      ImageSize = 16
	ImageSize32      ImageSize = 32
	ImageSize64      ImageSize = 64
	ImageSize128     ImageSize = 128
	ImageSize256     ImageSize = 256
	ImageSize512     ImageSize = 512
	ImageSize1024    ImageSize = 1024
	ImageSize2048    ImageSize = 2048
	ImageSize4096    ImageSize = 4096
)

// Is returns true if this ImageSize matches the provided one.
func (v ImageSize) Is(other ImageSize) bool {
	return v == other
}

// ImageFormat defines all possible image formats supported by Discord endpoints.
type ImageFormat string

const (
	ImageFormatDefault ImageFormat = ".gif"
	ImageFormatPNG     ImageFormat = ".png"
	ImageFormatJPEG    ImageFormat = ".jpeg"
	ImageFormatWebP    ImageFormat = ".webp"
	ImageFormatGIF     ImageFormat = ".gif"
	ImageFormatAVIF    ImageFormat = ".avif"
	ImageFormatLottie  ImageFormat = ".json"
)

// Is returns true if this ImageFormat matches the provided one.
func (v ImageFormat) Is(other ImageFormat) bool {
	return v == other
}

// ImageConfig holds configuration for image format and size.
type ImageConfig struct {
	Format ImageFormat
	Size   ImageSize
}

// isAnimatedHash checks if a hash represents an animated asset (starts with "a_").
func isAnimatedHash(hash string) bool {
	return len(hash) >= 2 && hash[:2] == "a_"
}

// buildImageURL constructs a URL for Discord assets with fallback logic.
func buildImageURL(baseURL, path, hash string, config ImageConfig, allowedFormats [5]ImageFormat) string {
	allowed := slices.Contains(allowedFormats[:], config.Format)
	animatedHash := isAnimatedHash(hash)

	if !allowed || (config.Format == ImageFormatGIF && !animatedHash) {
		config.Format = allowedFormats[0]
	}

	url := baseURL + path + string(config.Format)

	if config.Size > 0 {
		url += "?" + "size=" + strconv.Itoa(int(config.Size))
	}

	if config.Format == ImageFormatWebP && animatedHash {
		if config.Size > 0 {
			url += "&animated=true"
		} else {
			url += "?animated=true"
		}
	}
	return url
}

/***********************
 *        Emoji        *
 ***********************/

func EmojiURL(emojiID Snowflake, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatGIF, ImageFormatWebP, ImageFormatJPEG, ImageFormatAVIF}
	path := "emojis/" + emojiID.String()
	return buildImageURL(ImageBaseURL, path, "", config, allowedFormats)
}

/***********************
 *        Guild        *
 ***********************/

func GuildIconURL(guildID Snowflake, iconHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatGIF, ImageFormatWebP, ImageFormatJPEG}
	path := "icons/" + guildID.String() + "/" + iconHash
	return buildImageURL(ImageBaseURL, path, iconHash, config, allowedFormats)
}

func GuildSplashURL(guildID Snowflake, splashHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "splashes/" + guildID.String() + "/" + splashHash
	return buildImageURL(ImageBaseURL, path, splashHash, config, allowedFormats)
}

func GuildDiscoverySplashURL(guildID Snowflake, discoverySplashHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "discovery-splashes/" + guildID.String() + "/" + discoverySplashHash
	return buildImageURL(ImageBaseURL, path, discoverySplashHash, config, allowedFormats)
}

func GuildBannerURL(guildID Snowflake, bannerHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatGIF, ImageFormatWebP}
	path := "banners/" + guildID.String() + "/" + bannerHash
	return buildImageURL(ImageBaseURL, path, bannerHash, config, allowedFormats)
}

func GuildTagBadgeURL(guildID Snowflake, badgeHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "guild-tag-badges/" + guildID.String() + "/" + badgeHash
	return buildImageURL(ImageBaseURL, path, badgeHash, config, allowedFormats)
}

func GuildScheduledEventCoverURL(eventID Snowflake, coverHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "guild-events/" + eventID.String() + "/" + coverHash
	return buildImageURL(ImageBaseURL, path, coverHash, config, allowedFormats)
}

/***********************
 *         User        *
 ***********************/

func DefaultUserAvatarURL(index int) string {
	config := ImageConfig{Format: "", Size: 0}
	allowedFormats := [5]ImageFormat{ImageFormatPNG}
	path := "embed/avatars/" + strconv.Itoa(index)
	return buildImageURL(ImageBaseURL, path, "", config, allowedFormats)
}

func UserAvatarURL(userID Snowflake, avatarHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatGIF, ImageFormatWebP, ImageFormatJPEG}
	path := "avatars/" + userID.String() + "/" + avatarHash
	return buildImageURL(ImageBaseURL, path, avatarHash, config, allowedFormats)
}

func UserBannerURL(userID Snowflake, bannerHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatGIF, ImageFormatWebP, ImageFormatJPEG}
	path := "banners/" + userID.String() + "/" + bannerHash
	return buildImageURL(ImageBaseURL, path, bannerHash, config, allowedFormats)
}

/***********************
 *      Application    *
 ***********************/

func ApplicationIconURL(appID Snowflake, iconHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "app-icons/" + appID.String() + "/" + iconHash
	return buildImageURL(ImageBaseURL, path, iconHash, config, allowedFormats)
}

func ApplicationCoverURL(appID Snowflake, coverHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "app-icons/" + appID.String() + "/" + coverHash
	return buildImageURL(ImageBaseURL, path, coverHash, config, allowedFormats)
}

func ApplicationAssetURL(appID, assetID Snowflake, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "app-assets/" + appID.String() + "/" + assetID.String()
	return buildImageURL(ImageBaseURL, path, "", config, allowedFormats)
}

func AchievementIconURL(appID, achID Snowflake, iconHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "app-assets/" + appID.String() + "/achievements/" + achID.String() + "/icons/" + iconHash
	return buildImageURL(ImageBaseURL, path, iconHash, config, allowedFormats)
}

func StorePageAssetURL(appID, assetID Snowflake, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "app-assets/" + appID.String() + "/store/" + assetID.String()
	return buildImageURL(ImageBaseURL, path, "", config, allowedFormats)
}

/***********************
 *       Sticker       *
 ***********************/

func StickerURL(stickerID Snowflake, format ImageFormat) string {
	config := ImageConfig{Format: format}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatGIF, ImageFormatLottie}
	base := ImageBaseURL
	if config.Format == ImageFormatGIF {
		base = MediaBaseURL
	}
	path := "stickers/" + stickerID.String()
	return buildImageURL(base, path, "", config, allowedFormats)
}

func StickerPackBannerURL(stickerPackBannerAssetID Snowflake, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "app-assets/710982414301790216/store/" + stickerPackBannerAssetID.String()
	return buildImageURL(ImageBaseURL, path, "", config, allowedFormats)
}

/***********************
 *     Guild Member    *
 ***********************/

func GuildMemberAvatarURL(guildID, userID Snowflake, avatarHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatGIF, ImageFormatWebP, ImageFormatJPEG}
	path := "guilds/" + guildID.String() + "/users/" + userID.String() + "/avatars/" + avatarHash
	return buildImageURL(ImageBaseURL, path, avatarHash, config, allowedFormats)
}

func GuildMemberBannerURL(guildID, userID Snowflake, bannerHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatGIF, ImageFormatWebP, ImageFormatJPEG}
	path := "guilds/" + guildID.String() + "/users/" + userID.String() + "/banners/" + bannerHash
	return buildImageURL(ImageBaseURL, path, bannerHash, config, allowedFormats)
}

/***********************
 *      Guild Role     *
 ***********************/

func RoleIconURL(roleID Snowflake, iconHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "role-icons/" + roleID.String() + "/" + iconHash
	return buildImageURL(ImageBaseURL, path, iconHash, config, allowedFormats)
}

/***********************
 *  Avatar Decoration  *
 ***********************/

func AvatarDecorationURL(asset string, size ImageSize) string {
	config := ImageConfig{Format: "", Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG}
	path := "avatar-decoration-presets/" + asset
	return buildImageURL(ImageBaseURL, path, asset, config, allowedFormats)
}

/***********************
 *        Team         *
 ***********************/

func TeamIconURL(teamID Snowflake, iconHash string, format ImageFormat, size ImageSize) string {
	config := ImageConfig{Format: format, Size: size}
	allowedFormats := [5]ImageFormat{ImageFormatPNG, ImageFormatWebP, ImageFormatJPEG}
	path := "team-icons/" + teamID.String() + "/" + iconHash
	return buildImageURL(ImageBaseURL, path, iconHash, config, allowedFormats)
}
