FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Dependências Python
COPY ../server/requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r /app/requirements.txt

# Código (layout src/)
COPY ../server/src/ /app/src
ENV PYTHONPATH=/app/src

# Usuário e diretório de dados
RUN useradd -m -u 1000 appuser && mkdir -p /data && chown -R appuser:appuser /app /data

# Copia o entrypoint já com permissão e dono corretos (evita RUN chmod)
COPY --chmod=0755 --chown=appuser:appuser ./entrypoint.sh /usr/local/bin/entrypoint.sh

USER appuser
EXPOSE 8040

CMD ["/usr/local/bin/entrypoint.sh"]
