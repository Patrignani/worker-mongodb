package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/storage"
)

// Configuração
const (
	bucketName   = "SEU_BUCKET"
	daysToDelete = 30 // Arquivos mais antigos que 30 dias serão excluídos
)

func main() {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Erro ao criar cliente GCS: %v", err)
	}
	defer client.Close()

	// Calcula a data limite para exclusão
	cutoffDate := time.Now().AddDate(0, 0, -daysToDelete)

	log.Printf("Excluindo arquivos do bucket %s anteriores a %s...\n", bucketName, cutoffDate.Format("2006-01-02"))

	// Lista e deleta arquivos
	if err := deleteOldFiles(ctx, client, bucketName, cutoffDate); err != nil {
		log.Fatalf("Erro ao excluir arquivos antigos: %v", err)
	}

	log.Println("Processo de limpeza concluído!")
}

// deleteOldFiles exclui arquivos mais antigos que `cutoffDate`
func deleteOldFiles(ctx context.Context, client *storage.Client, bucketName string, cutoffDate time.Time) error {
	bucket := client.Bucket(bucketName)
	it := bucket.Objects(ctx, nil)

	for {
		objAttrs, err := it.Next()
		if err != nil {
			if err.Error() == "iterator.Done" {
				break
			}
			return fmt.Errorf("erro ao listar objetos: %v", err)
		}

		// Compara a data do arquivo com a data limite
		if objAttrs.Updated.Before(cutoffDate) {
			log.Printf("Excluindo: %s (Última modificação: %s)\n", objAttrs.Name, objAttrs.Updated)

			// Exclui o objeto
			if err := bucket.Object(objAttrs.Name).Delete(ctx); err != nil {
				log.Printf("Erro ao excluir %s: %v\n", objAttrs.Name, err)
			} else {
				log.Printf("Arquivo %s excluído com sucesso.\n", objAttrs.Name)
			}
		}
	}

	return nil
}
