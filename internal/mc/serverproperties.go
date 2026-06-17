package mc

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ServerProperties struct {
	ServerName         string `json:"server_name"`
	ServerPort         int    `json:"server_port"`
	MaxPlayers         int    `json:"max_players"`
	MOTD               string `json:"motd"`
	LevelName          string `json:"level_name"`
	LevelSeed          string `json:"level_seed"`
	Gamemode           string `json:"gamemode"`
	Difficulty         string `json:"difficulty"`
	AllowNether        bool   `json:"allow_nether"`
	SpawnMonsters      bool   `json:"spawn_monsters"`
	SpawnAnimals       bool   `json:"spawn_animals"`
	SpawnNpcs          bool   `json:"spawn_npcs"`
	OnlineMode         bool   `json:"online_mode"`
	PVP                bool   `json:"pvp"`
	Hardcore           bool   `json:"hardcore"`
	WhiteList          bool   `json:"white_list"`
	ViewDistance       int    `json:"view_distance"`
	SimulationDistance int    `json:"simulation_distance"`
	LevelType          string `json:"level_type"`
}

func DefaultProps() ServerProperties {
	return ServerProperties{
		ServerName:         "A Minecraft Server",
		MOTD:               "A Minecraft Server",
		MaxPlayers:         20,
		Gamemode:           "survival",
		Difficulty:         "normal",
		AllowNether:        true,
		SpawnMonsters:      true,
		SpawnAnimals:       true,
		SpawnNpcs:          true,
		OnlineMode:         true,
		PVP:                true,
		Hardcore:           false,
		WhiteList:          false,
		ViewDistance:       10,
		SimulationDistance: 10,
		LevelType:          "minecraft:normal",
		ServerPort:         25565,
	}
}

func ReadServerProperties(dir string) (ServerProperties, error) {
	path := filepath.Join(dir, "server.properties")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultProps(), nil
		}
		return DefaultProps(), err
	}

	props := DefaultProps()
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "server-name":
			props.ServerName = val
		case "server-port":
			props.ServerPort, _ = strconv.Atoi(val)
		case "motd":
			props.MOTD = val
		case "max-players":
			props.MaxPlayers, _ = strconv.Atoi(val)
		case "level-name":
			props.LevelName = val
		case "level-seed":
			props.LevelSeed = val
		case "gamemode":
			props.Gamemode = val
		case "difficulty":
			props.Difficulty = val
		case "allow-nether":
			props.AllowNether = val == "true"
		case "spawn-monsters":
			props.SpawnMonsters = val == "true"
		case "spawn-animals":
			props.SpawnAnimals = val == "true"
		case "spawn-npcs":
			props.SpawnNpcs = val == "true"
		case "online-mode":
			props.OnlineMode = val == "true"
		case "pvp":
			props.PVP = val == "true"
		case "hardcore":
			props.Hardcore = val == "true"
		case "white-list":
			props.WhiteList = val == "true"
		case "view-distance":
			props.ViewDistance, _ = strconv.Atoi(val)
		case "simulation-distance":
			props.SimulationDistance, _ = strconv.Atoi(val)
		case "level-type":
			props.LevelType = val
		}
	}

	return props, nil
}

func WriteServerProperties(dir string, props ServerProperties) error {
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "server.properties")

	var sb strings.Builder
	sb.WriteString("# Minecraft server properties (managed by craftctl)\n")
	sb.WriteString("server-name=" + props.ServerName + "\n")
	sb.WriteString("server-port=" + strconv.Itoa(props.ServerPort) + "\n")
	sb.WriteString("motd=" + props.MOTD + "\n")
	sb.WriteString("max-players=" + strconv.Itoa(props.MaxPlayers) + "\n")
	sb.WriteString("level-name=" + props.LevelName + "\n")
	sb.WriteString("level-seed=" + props.LevelSeed + "\n")
	sb.WriteString("gamemode=" + props.Gamemode + "\n")
	sb.WriteString("difficulty=" + props.Difficulty + "\n")
	sb.WriteString("allow-nether=" + boolStr(props.AllowNether) + "\n")
	sb.WriteString("spawn-monsters=" + boolStr(props.SpawnMonsters) + "\n")
	sb.WriteString("spawn-animals=" + boolStr(props.SpawnAnimals) + "\n")
	sb.WriteString("spawn-npcs=" + boolStr(props.SpawnNpcs) + "\n")
	sb.WriteString("online-mode=" + boolStr(props.OnlineMode) + "\n")
	sb.WriteString("pvp=" + boolStr(props.PVP) + "\n")
	sb.WriteString("hardcore=" + boolStr(props.Hardcore) + "\n")
	sb.WriteString("white-list=" + boolStr(props.WhiteList) + "\n")
	sb.WriteString("view-distance=" + strconv.Itoa(props.ViewDistance) + "\n")
	sb.WriteString("simulation-distance=" + strconv.Itoa(props.SimulationDistance) + "\n")
	sb.WriteString("level-type=" + props.LevelType + "\n")

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
