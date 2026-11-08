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
	POSTGRES_SMOC     int
	POSTGRES_SMIC     int
	DB_URL            string
	POSTGRES_US_SMOC  int
	POSTGRES_US_SMIC  int
	REPO_URL          string
	TG_CHANNEL_URL    string
	TG_CHANNEL_ID     int64
	MAX_BRO           int
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

	db_url := os.Getenv("DB_URL")
	if db_url == "" {
		return nil, fmt.Errorf("error when getting db mus url")
	}
	db_smoc := os.Getenv("POSTGRES_SMOC")
	if db_smoc == "" {
		return nil, fmt.Errorf("error whem getting db mus smoc")
	}
	db_smic := os.Getenv("POSTGRES_SMIC")
	if db_smic == "" {
		return nil, fmt.Errorf("error whem getting db mus smic")
	}
	int_db_smoc, err := strconv.Atoi(db_smoc)
	if err != nil {
		return nil, err
	}
	int_db_smic, err := strconv.Atoi(db_smic)
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

	mxb := os.Getenv("MAX_BRO")
	if mxb == "" {
		return nil, fmt.Errorf("errror when get max brouser")
	}
	int_mxb, err := strconv.Atoi(mxb)
	if err != nil {
		return nil, err
	}

	tg_id := os.Getenv("TG_CHANNEL_ID")
	if tg_id == "" {
		return nil, fmt.Errorf("error when getting tg chanal id")
	}
	int_th_id, err := strconv.ParseInt(tg_id, 10, 64)
	if err != nil {
		return nil, err
	}

	return &Config{
		BOT_TOKEN:         bot_token,
		COMMUN_PORT:       commun_port,
		CAMOUFOX_PORT:     camoufox_port,
		HEALTH_CHECK_PORT: health_check_port,
		DB_URL:            db_url,
		POSTGRES_SMOC:     int_db_smoc,
		POSTGRES_SMIC:     int_db_smic,
		REPO_URL:          repo_url,
		TG_CHANNEL_URL:    tg_channel_url,
		TG_CHANNEL_ID:     int_th_id,
		MAX_BRO:           int_mxb,
	}, nil
}
