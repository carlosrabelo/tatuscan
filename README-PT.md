# TatuScan - Sistema Distribuído de Inventário de Máquinas

TatuScan é um sistema distribuído de inventário de máquinas com um cliente leve em Go e um servidor Python Flask para coleta e monitoramento de informações de máquinas em redes.

## Visão Geral

TatuScan consiste em dois componentes principais:
- **Cliente**: Agente leve em Go que coleta informações do sistema (Windows/Linux)
- **Servidor**: Servidor API baseado em Flask com suporte a SQLite/PostgreSQL/MySQL

O sistema coleta dados das máquinas a cada 60 segundos e fornece um painel web para visualização e relatórios.

## Arquitetura

```
┌─────────────────┐    HTTP POST JSON    ┌─────────────────┐
│  Cliente Go     │ ───────────────────→ │ Servidor Flask  │
│  (Windows/Linux)│    /api/machines     │   + Banco de    │
└─────────────────┘                      │     Dados       │
         │                               └─────────────────┘
         │ intervalo 60s                         │
         │                              Painel Web (Dashboard)
         └───────────────────────────────────────┘
```

## Recursos

### Cliente (Go)
- **Multiplataforma**: Windows 7-11, Linux (distribuições modernas)
- **Leve**: Consumo mínimo de recursos
- **Integração com serviços**: Suporte a systemd (Linux) e Windows Services
- **Identificação segura**: ID da máquina baseado em endereços MAC físicos (SHA-256)
- **Intervalo configurável**: Frequência de coleta ajustável
- **Filtragem robusta**: Exclui interfaces virtuais/cloud automaticamente

### Servidor (Python Flask)
- **API RESTful**: Endpoints JSON limpos para ingestão de dados
- **Múltiplos bancos de dados**: Suporte a SQLite (padrão), PostgreSQL, MySQL
- **Painel web**: Interface HTML para visualização de inventário e relatórios
- **Suporte a fusos horários**: Fuso horário configurável (padrão: America/Cuiaba)
- **Pronto para Docker**: Suporte a implantação em contêineres
- **Pronto para produção**: Integração com Supervisord e Nginx

## Início Rápido

### Pré-requisitos
- **Go 1.23+** (para o cliente)
- **Python 3.10+** (para o servidor)
- **Docker** (opcional, para implantação em contêineres)

### Instalação

1. **Clone o repositório**:
   ```bash
   git clone <URL_DO_REPOSITORIO>
   cd tatuscan
   ```

2. **Compile o cliente**:
   ```bash
   make client-build
   ```

3. **Configure o servidor**:
   ```bash
   cd server
   python3 -m venv .venv
   source .venv/bin/activate
   pip install -r requirements.txt
   cd ..
   ```

4. **Inicie o servidor**:
   ```bash
   make server-start
   ```

5. **Configure e execute o cliente**:
   ```bash
   export TATUSCAN_URL=http://localhost:8040
   ./client/tatuscan
   ```

## Estrutura do Projeto

```
tatuscan/
├── bin/                   # Binários compilados
│   ├── linux/            # Executáveis Linux
│   └── windows/          # Executáveis Windows
├── client/                # Aplicação cliente em Go
│   ├── cmd/tatuscan/     # Ponto de entrada principal
│   ├── internal/         # Pacotes internos
│   ├── tools/            # Scripts utilitários
│   ├── tatuscan.wxs      # Configuração instalador Windows
│   ├── .env.example      # Template de variáveis de ambiente
│   └── go.mod/go.sum     # Dependências Go
├── server/               # Servidor Flask em Python
│   ├── tatuscan/         # Pacote principal da aplicação
│   │   ├── blueprint/    # Blueprints Flask (api, home, report, charts)
│   │   ├── config/       # Módulo de configuração
│   │   ├── errors/       # Tratadores de erro
│   │   ├── logging/      # Configuração de logs
│   │   ├── models/       # Modelos do banco de dados
│   │   ├── services/     # Camada de lógica de negócio
│   │   ├── utils/        # Utilitários compartilhados
│   │   └── templates/    # Templates HTML
│   ├── scripts/          # Scripts utilitários do banco
│   ├── requirements.txt  # Dependências Python
│   ├── run.py            # Ponto de entrada desenvolvimento
│   ├── Dockerfile        # Configuração do container
│   ├── docker-compose.yml # Compose para desenvolvimento local
│   └── .env.example      # Template de variáveis de ambiente
├── deploy/               # Configurações de deployment
│   ├── docker/           # Setup Docker para produção
│   ├── k8s/              # Manifestos Kubernetes
│   ├── systemd/          # Serviços systemd Linux
│   └── README.md         # Guia de deployment
├── scripts/              # Scripts de build e deployment
│   ├── client-build.sh   # Script de build do cliente
│   ├── server-*.sh       # Scripts de gerenciamento do servidor
│   └── clean*.sh         # Scripts de limpeza
├── Makefile             # Makefile principal do projeto
├── README.md            # Documentação (inglês)
└── README-PT.md         # Este arquivo (português)
```

## Configuração

### Configuração do Cliente

Crie o arquivo `.env` no diretório `client/`:

```bash
# URL do servidor (obrigatório)
TATUSCAN_URL=http://localhost:8040

# Intervalo de coleta (opcional, padrão: 60s)
TATUSCAN_INTERVAL=60s

# Nível de log (opcional, padrão: warn)
TATUSCAN_LOG_LEVEL=warn
```

### Configuração do Servidor

Crie o arquivo `.env` no diretório `server/`:

```bash
# Porta do servidor (padrão: 8040)
TATUSCAN_PORT=8040

# Configuração do banco de dados
# SQLite (padrão)
TATUSCAN_DB_DIR=/data
TATUSCAN_DB_FILE=tatuscan.db

# PostgreSQL (opcional)
# SQLALCHEMY_DATABASE_URI=postgresql+psycopg2://user:pass@host:5432/dbname

# MySQL (opcional)
# SQLALCHEMY_DATABASE_URI=mysql+pymysql://user:pass@host:3306/dbname

# Fuso horário
TIMEZONE=America/Cuiaba

# Chave secreta do Flask
SECRET_KEY=change-me-in-production
```

### Variáveis de Ambiente

Principais variáveis lidas pela aplicação:

- `TATUSCAN_PORT`: Porta interna do app. Se ausente, usa `PORT` como fallback. Padrão: `8040`
- `PORT`: Fallback para porta quando `TATUSCAN_PORT` não está definido
- `SQLALCHEMY_DATABASE_URI`: URI completa do banco de dados (tem prioridade se definida)
- `TATUSCAN_DB_DIR`: Diretório do SQLite quando `SQLALCHEMY_DATABASE_URI` não está definido. Padrão: `/data`
- `TATUSCAN_DB_FILE`: Nome do arquivo SQLite. Padrão: `tatuscan.db`
- `TIMEZONE`: Fuso horário para exibição. Padrão: `America/Cuiaba`
- `SECRET_KEY`: Chave secreta do Flask. Padrão: `dev` (mudar em produção)

## Uso

### Compilando o Projeto

```bash
# Compilar cliente para a plataforma atual
make client-build

# Compilar cliente para todas as plataformas
make client-build-all

# Compilar imagem Docker do servidor
make server-build
```

### Executando as Aplicações

#### Desenvolvimento

**Servidor:**
```bash
cd server
source .venv/bin/activate
./run.py
```

Acesse:
- Dashboard: `http://localhost:8040/`
- Relatório: `http://localhost:8040/report`
- Gráficos: `http://localhost:8040/charts`
- Endpoint da API: `http://localhost:8040/api/machines`

**Cliente:**
```bash
# Executar cliente (coleta única)
make client-run

# Executar cliente como daemon
make client-daemon
```

#### Implantação com Docker

Use os alvos do Makefile para gerenciar o contêiner:

```bash
make server-start   # Iniciar com docker compose up -d
make server-stop    # Parar com docker compose down
make server-restart # Reiniciar apenas o serviço do app
make server-logs    # Seguir os logs
make server-ps      # Status dos contêineres
```

**Notas:**
- O `docker-compose.yml` lê automaticamente o `.env`
- Mapeamento de porta: `${HOST_PORT:-8040}:${TATUSCAN_PORT:-8040}`

**Sobrescrever a porta do host (HOST_PORT):**

Opção 1 — via `.env` (recomendado):
```env
# .env
HOST_PORT=18040       # Porta exposta no host
TATUSCAN_PORT=8040    # Porta interna do app (não mudar normalmente)
```

Então:
```bash
make server-start
# Acesse em http://localhost:18040
```

Opção 2 — ad-hoc via shell:
```bash
HOST_PORT=18040 make server-start
```

### Implantação em Produção

1. **Configuração do Servidor**:
   ```bash
   # Criar diretórios
   sudo mkdir -p /opt/tatuscan
   sudo mkdir -p /var/lib/tatuscan

   # Copiar arquivos para o diretório de produção
   sudo cp -r server/* /opt/tatuscan/

   # Configurar permissões
   sudo chown -R www-data:www-data /opt/tatuscan
   sudo chown -R www-data:www-data /var/lib/tatuscan
   sudo chmod 750 /opt/tatuscan
   sudo chmod 750 /var/lib/tatuscan

   # Configurar supervisord
   sudo cp deploy/tatuscan.conf /etc/supervisor/conf.d/
   sudo supervisorctl reread
   sudo supervisorctl update
   sudo supervisorctl start tatuscan
   ```

2. **Instalação do Cliente**:
   ```bash
   # Instalar como serviço (Linux)
   sudo make client-install

   # Instalar como serviço (Windows)
   ./bin/tatuscan-windows-amd64.exe install
   ```

3. **Configuração do Nginx**:
   ```bash
   # Copiar configuração do Nginx
   sudo cp deploy/nginx/tatuscan.conf /etc/nginx/sites-available/
   sudo ln -s /etc/nginx/sites-available/tatuscan.conf /etc/nginx/sites-enabled/
   sudo nginx -t
   sudo systemctl reload nginx
   ```

### Comandos de Desenvolvimento

```bash
# Formatar código Go
make client-fmt

# Executar testes Go
make client-test

# Executar testes Python
make server-test

# Fazer lint do código
make lint

# Limpar artefatos de compilação
make clean
```

## Endpoints da API

### GET /api/machines
Lista todas as máquinas registradas.

**Resposta:**
```json
{
  "items": [
    {
      "machine_id": "sha256-hash",
      "hostname": "server-01",
      "ip": "192.168.1.100",
      "os": "linux",
      "os_version": "Ubuntu 22.04",
      "cpu_percent": 15.5,
      "memory_total_mb": 8192,
      "memory_used_mb": 4096,
      "created_at": "2025-01-01T12:00:00-04:00",
      "updated_at": "2025-01-01T12:30:00-04:00"
    }
  ],
  "count": 1
}
```

### POST /api/machines
Recebe dados dos clientes TatuScan.

**Corpo da Requisição:**
```json
{
  "machine_id": "sha256-hash",
  "hostname": "server-01",
  "ip": "192.168.1.100",
  "os": "linux",
  "os_version": "Ubuntu 22.04",
  "cpu_percent": 15.5,
  "memory_total_mb": 8192,
  "memory_used_mb": 4096,
  "timestamp": "2025-01-01T12:30:00-04:00"
}
```

### GET /api/health
Endpoint de verificação de saúde.

**Resposta:**
```json
{
  "status": "healthy"
}
```

## Dados Coletados

O cliente coleta as seguintes informações:

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `machine_id` | string | Hash SHA-256 dos endereços MAC físicos |
| `hostname` | string | Nome do host da máquina |
| `ip` | string | Endereço IPv4 principal |
| `os` | string | Sistema operacional (linux/windows/darwin) |
| `os_version` | string | Versão do SO legível por humanos |
| `cpu_percent` | float | Porcentagem de uso da CPU |
| `memory_total_mb` | integer | Memória total em MB |
| `memory_used_mb` | integer | Memória usada em MB |
| `timestamp` | string | Timestamp ISO 8601 |

## Estrutura do Banco de Dados

A tabela `Inventory` contém:

- `machine_id`: Chave primária (hash SHA-256, único por máquina)
- `hostname`: Nome do host
- `ip`: Endereço IP
- `os`: Sistema operacional
- `os_version`: Versão do sistema operacional
- `cpu_percent`: Porcentagem de uso da CPU
- `memory_total_mb`: Memória total em MB
- `memory_used_mb`: Memória usada em MB
- `created_at`: Data de criação (UTC)
- `updated_at`: Data da última atualização (fuso horário configurado)

Para inspecionar o banco de dados:

```bash
# Desenvolvimento (quando usar SQLite local)
sqlite3 /caminho/para/tatuscan/tatuscan.db

# Produção
sqlite3 /var/lib/tatuscan/tatuscan.db

# Consulta
SELECT * FROM inventory;
```

## Usando Banco de Dados Externo (Postgres/MySQL)

Para usar PostgreSQL ou MySQL em vez de SQLite:

1. Defina `SQLALCHEMY_DATABASE_URI` no `.env`:

```env
# PostgreSQL (driver psycopg2)
SQLALCHEMY_DATABASE_URI=postgresql+psycopg2://user:pass@host:5432/dbname

# MySQL/MariaDB (driver PyMySQL)
SQLALCHEMY_DATABASE_URI=mysql+pymysql://user:pass@host:3306/dbname
```

2. Drivers já incluídos no `requirements.txt`:
   - PostgreSQL: `psycopg2-binary`
   - MySQL/MariaDB: `PyMySQL`

Se não for usar banco de dados externo, você pode removê-los do `requirements.txt` para uma imagem mais enxuta.

### Sobre o volume `/data` no Docker

Quando `SQLALCHEMY_DATABASE_URI` está definido para Postgres/MySQL, o volume mapeado em `/data` não é usado, mas também não atrapalha. Se quiser removê-lo:

**Opção 1** — Editar `docker-compose.yml` e comentar a linha do volume:
```yaml
services:
  tatuscan:
    # ...
    # volumes:
    #   - tatuscan_data:/data
```

**Opção 2** — Usar um arquivo de override para não mapear `/data`:

Crie `docker-compose.no-sqlite.yml`:
```yaml
services:
  tatuscan:
    volumes: []
```

E inicie com:
```bash
docker compose -f docker-compose.yml -f docker-compose.no-sqlite.yml up -d
```

## Testando com o Cliente TatuScan

1. **Compile o cliente**:
   ```bash
   cd client
   go build -o tatuscan cmd/tatuscan/main.go
   ```

2. **Configure a variável de ambiente**:
   ```bash
   # Desenvolvimento
   export TATUSCAN_URL=http://localhost:8040

   # Produção
   export TATUSCAN_URL=http://tatuscan.example.com
   ```

3. **Execute o cliente**:
   ```bash
   ./tatuscan
   ```

4. **Verifique os dados no relatório**:
   ```bash
   curl http://localhost:8040/report
   ```

   Ou acesse via navegador.

## Considerações de Segurança

- **ID da Máquina**: Baseado apenas em endereços MAC físicos (exclui interfaces virtuais)
- **HTTPS**: Use HTTPS em ambientes de produção
- **Autenticação**: Considere adicionar autenticação de API para produção
- **Firewall**: Configure regras de firewall apropriadas
- **Segredos**: Nunca faça commit de arquivos `.env` ou segredos no controle de versão
- **Banco de Dados**: Use senhas fortes e restrinja acesso aos servidores de banco de dados
- **Permissões de Arquivos**: Configure permissões apropriadas em arquivos de configuração e bancos de dados

## Solução de Problemas

### Problemas do Cliente

**Problema**: "No valid physical network interface found"
- **Solução**: Verifique se a máquina possui interfaces de rede físicas conectadas

**Problema**: "Environment variable TATUSCAN_URL not defined"
- **Solução**: Defina a variável de ambiente TATUSCAN_URL

**Problema**: Connection refused
- **Solução**: Verifique se o servidor está em execução e se a URL está correta

### Problemas do Servidor

**Problema**: Erros de conexão com o banco de dados
- **Solução**: Verifique a configuração e permissões do banco de dados

**Problema**: Porta já em uso
- **Solução**: Altere TATUSCAN_PORT ou pare os serviços conflitantes

**Problema**: Permissão negada em `/var/lib/tatuscan`
- **Solução**: Verifique propriedade e permissões (www-data:www-data, 750)

**Problema**: Supervisord não inicia
- **Solução**: Verifique os logs em `/var/log/tatuscan.err.log` e `/var/log/tatuscan.out.log`

## Contribuindo

1. Faça um fork do repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas alterações (`git commit -m 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

## Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## Suporte

Para suporte e dúvidas:
- Crie uma issue no repositório
- Contato: contato@carlosrabelo.com.br
