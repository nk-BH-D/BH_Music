package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	BOT_TOKEN         string
	COMMUN_PORT       string
	CAMOUFOX_PORT     string
	HEALTH_CHECK_PORT string
	DB_MUS_URL        string
	POSTGRES_MUS_SMOC int
	POSTGRES_MUS_SMIC int
	DB_US_URL         string
	POSTGRES_US_SMOC  int
	POSTGRES_US_SMIC  int
	REPO_URL          string
	TG_CHANNEL_URL    string
}

func Loader() (*Config, error) {

	bot_token := os.Getenv("BOT_TOKEN")
	if bot_token == "" {
		return nil, fmt.Errorf("error whem getting bot token")
	}

	commun_port := os.Getenv("COMMUN_PORT")
	if commun_port == "" {
		return nil, fmt.Errorf("error whem get commun port")
	}

	camoufox_port := os.Getenv("CAMOUFOX_PORT")
	if camoufox_port == "" {
		return nil, fmt.Errorf("error whem get camoufox port")
	}

	health_check_port := os.Getenv("HEALTH_CHECK_PORT")
	if health_check_port == "" {
		return nil, fmt.Errorf("error whem getting health check port")
	}

	db_mus_url := os.Getenv("DB_MUS_URL")
	if db_mus_url == "" {
		return nil, fmt.Errorf("error when getting db mus url")
	}
	db_mus_smoc := os.Getenv("POSTGRES_MUS_SMOC")
	if db_mus_smoc == "" {
		return nil, fmt.Errorf("error whem getting db mus smoc")
	}
	db_mus_smic := os.Getenv("POSTGRES_MUS_SMIC")
	if db_mus_smic == "" {
		return nil, fmt.Errorf("error whem getting db mus smic")
	}
	int_db_mus_smoc, err := strconv.Atoi(db_mus_smoc)
	if err != nil {
		return nil, err
	}
	int_db_mus_smic, err := strconv.Atoi(db_mus_smic)
	if err != nil {
		return nil, err
	}

	db_us_url := os.Getenv("DB_US_URL")
	if db_us_url == "" {
		return nil, fmt.Errorf("error whem getting db users url")
	}
	db_us_smoc := os.Getenv("POSTGRES_US_SMOC")
	if db_us_smoc == "" {
		return nil, fmt.Errorf("error whem getting db users smoc")
	}
	db_us_smic := os.Getenv("POSTGRES_US_SMIC")
	if db_us_smic == "" {
		return nil, fmt.Errorf("error whem getting db users smic")
	}
	int_db_us_smoc, err := strconv.Atoi(db_us_smoc)
	if err != nil {
		return nil, err
	}
	int_db_us_smic, err := strconv.Atoi(db_us_smic)
	if err != nil {
		return nil, err
	}

	repo_url := os.Getenv("REPO_URL")
	if repo_url == "" {
		return nil, fmt.Errorf("error whem getting repo url")
	}

	tg_channel_url := os.Getenv("TG_CHANNEL_URL")
	if tg_channel_url == "" {
		return nil, fmt.Errorf("error whem getting tg channel url")
	}

	return &Config{
		BOT_TOKEN:         bot_token,
		COMMUN_PORT:       commun_port,
		CAMOUFOX_PORT:     camoufox_port,
		HEALTH_CHECK_PORT: health_check_port,
		DB_MUS_URL:        db_mus_url,
		DB_US_URL:         db_us_url,
		POSTGRES_MUS_SMOC: int_db_mus_smoc,
		POSTGRES_MUS_SMIC: int_db_mus_smic,
		POSTGRES_US_SMOC:  int_db_us_smoc,
		POSTGRES_US_SMIC:  int_db_us_smic,
		REPO_URL:          repo_url,
		TG_CHANNEL_URL:    tg_channel_url,
	}, nil
}
