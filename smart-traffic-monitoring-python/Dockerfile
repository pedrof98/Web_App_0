FROM python:3.11-slim

# working directory
WORKDIR /app

# Command to instal system dependencies goes here:
RUN apt-get update && apt-get install -y \
libpq-dev \
netcat-traditional \
&& rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

# Install copied dependencies
RUN pip install --no-cache-dir -r requirements.txt

COPY wait-for-kafka.sh /wait-for-kafka.sh
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /wait-for-kafka.sh
RUN chmod +x /entrypoint.sh
# Now only copy the minimal set of folders you actually need.
COPY app/ /app/app/        
COPY alembic/ /app/alembic/
COPY alembic.ini /app/
COPY tests/ /app/tests/
#COPY docker-compose.yml /app/   

# Expose port for the api
EXPOSE 8900

ENV PYTHONPATH=/app

# command to run on container start (usually the main.py script to run the whole app)
ENTRYPOINT ["/entrypoint.sh"]
CMD ["python3", "app/main.py"]
#CMD ["/wait-for-kafka.sh", "kafka", "9092", "python3", "app/main.py"]

# i will change this to run uvicorn in the future, after developing the app further


