docker build -t fineract-cbdc-india-connector:latest -f deployments/Dockerfile .
docker run -p 8080:8080 fineract-cbdc-india-connector:latest