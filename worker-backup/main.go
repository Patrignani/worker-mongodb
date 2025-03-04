package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Patrignani/worker-backup-mongodb/config"
)

func createMongoBackup() (string, error) {
	dateStr := time.Now().Format("2006-01-02")
	outputDir := fmt.Sprintf("%s/backup-%s", config.Env.BackupDir, dateStr)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("erro ao criar diretório de backup: %v", err)
	}

	cmd := exec.Command("mongodump",
		"--host", config.Env.MongoHost,
		"--port", config.Env.MongoPort,
		"--username", config.Env.MongoUser,
		"--password", config.Env.MongoPassword,
		"--db", config.Env.MongoDB,
		"--out", outputDir,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("erro ao executar mongodump: %v", err)
	}

	log.Println("Backup concluído:", outputDir)
	return outputDir, nil
}

func zipBackup(backupDir string) (string, error) {
	zipFile := backupDir + ".zip"

	cmd := exec.Command("zip", "-r", "-9", "-P", config.Env.ZipPassword, zipFile, backupDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("erro ao criar ZIP: %v", err)
	}

	log.Println("Backup compactado com senha:", zipFile)
	return zipFile, nil
}

func uploadToGCP(zipFilePath string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("erro ao criar cliente do GCP Storage: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(config.Env.GCPBucketName)
	object := bucket.Object(filepath.Base(zipFilePath))
	writer := object.NewWriter(ctx)

	file, err := os.Open(zipFilePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo para upload: %v", err)
	}
	defer file.Close()

	if _, err = io.Copy(writer, file); err != nil {
		return fmt.Errorf("erro ao fazer upload para o GCP Storage: %v", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("erro ao finalizar upload: %v", err)
	}

	log.Println("Backup enviado para o GCP Storage:", zipFilePath)
	return nil
}

func cleanUpBackup(backupDir, zipFile string) {
	log.Println("Apagando backup local...")
	if err := os.RemoveAll(backupDir); err != nil {
		log.Println("Erro ao apagar backup:", err)
	} else {
		log.Println("Backup apagado:", backupDir)
	}

	if err := os.Remove(zipFile); err != nil {
		log.Println("Erro ao apagar arquivo zipado:", err)
	} else {
		log.Println("Arquivo zipado apagado:", zipFile)
	}
}

func main() {
	for {
		log.Println("Iniciando backup do MongoDB...")

		backupPath, err := createMongoBackup()
		if err != nil {
			log.Fatal(err)
		}

		zipFile, err := zipBackup(backupPath)
		if err != nil {
			log.Fatal(err)
		}

		if err := uploadToGCP(zipFile); err != nil {
			log.Fatal("Erro ao fazer upload para GCP:", err)
		}

		cleanUpBackup(backupPath, zipFile)

		time.Sleep(time.Hour * 24)
	}
}
