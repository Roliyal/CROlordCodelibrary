FROM python:3.11-slim

WORKDIR /app

COPY python-svc/requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r /app/requirements.txt

COPY python-svc/app.py /app/app.py

COPY .env /app/.env
ENV DOTENV_PATH=/app/.env

EXPOSE 8081
CMD ["python", "/app/app.py"]
