FROM python:3.11-slim

# working directory
WORKDIR /app

# Command to instal system dependencies goes here:
RUN apt-get update && apt-get install -y \
gcc \
python3-dev \
libpq-dev

COPY requirements.txt .

# Install copied dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the app codebase
COPY . .

# Expose port for the api
EXPOSE 8900

ENV PYTHONPATH=/app

# command to run on container start (usually the main.py script to run the whole app)
CMD ["python3", "app/main.py"]

# i will change this to run uvicorn in the future, after developing the app further


