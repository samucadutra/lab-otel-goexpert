# Lab OpenTelemetry Go Expert

Este projeto demonstra a implementação de OpenTelemetry com Go, utilizando microsserviços para rastreamento distribuído.

## Requisitos

- Docker
- Docker Compose
- IDE compatível com arquivos HTTP (como VSCode ou Goland) para utilizar o arquivo de requisições

## Executando o Projeto

Para iniciar todos os serviços, execute o seguinte comando na raiz do projeto:

```bash
docker-compose up -d
```

Este comando iniciará:
- Coletor OpenTelemetry
- Zipkin (para visualização de traces)
- Aplicação Go principal (porta 8080)
- Segundo serviço Go (porta 8181)

## Endpoints Disponíveis

### Serviço A (Principal)

- **Endpoint:** `POST http://localhost:8080/weather/servico-a`
- **Corpo da requisição:** JSON com campo `cep` (CEP brasileiro)
- **Exemplo:**
  ```json
  {
    "cep": "12345678"
  }
  ```

### Serviço B

- **Endpoint:** `GET http://localhost:8080/weather/servico-b/{zipcode}`
- **Parâmetro:** `zipcode` - CEP brasileiro
- **Exemplo:** `http://localhost:8080/weather/servico-b/12345678`

### Métricas

- **Endpoint:** `GET http://localhost:8080/metrics`
- Retorna métricas no formato Prometheus

## Usando o Arquivo de Requisições HTTP

O projeto inclui um arquivo `initial_request.http` que pode ser usado para testar os endpoints facilmente.

Se você estiver usando VSCode:

1. Instale a extensão "REST Client" por Huachao Mao
2. Abra o arquivo `initial_request.http`
3. Clique em "Send Request" acima de cada definição de requisição

Se estiver usando JetBrains IDEs (IntelliJ, GoLand, etc.):

1. Abra o arquivo `initial_request.http`
2. Clique no botão de execução (play) ao lado de cada requisição

## Visualizando Traces

Para visualizar os traces gerados:

1. Acesse o Zipkin em `http://localhost:9411`
2. Utilize a interface para buscar e analisar os traces

## Configuração

O projeto utiliza variáveis de ambiente para configuração, definidas no arquivo `docker-compose.yaml`. 
As principais variáveis incluem:

- `WEATHER_API_KEY`: Chave para a API de previsão do tempo
- `WEB_SERVER_PORT`: Porta em que o servidor web será executado
- `OTEL_SERVICE_NAME`: Nome do serviço para rastreamento

## Solucionando Problemas

Se encontrar problemas ao executar o projeto:

1. Verifique se as portas 8080, 8181 e 9411 estão disponíveis
2. Confira os logs dos containers: `docker-compose logs -f`
3. Reinicie todos os serviços: `docker-compose restart`
