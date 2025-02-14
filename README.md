Backend Application with Pinecone

This backend application integrates Pinecone for vector storage and search capabilities using an LLM (Large Language Model). It allows users to upload PDFs, extract text from them, store embeddings in Pinecone, and perform search queries.

Features

Upsert PDFs: Extracts text from PDFs and stores them in Pinecone as vector embeddings.

Search API: Allows querying stored embeddings for similarity search using an LLM.

Tech Stack

Go/Python (Specify your backend language)

Pinecone for vector storage

LLM (e.g., OpenAI, Cohere, or another model) for embeddings

FastAPI/Flask/Gin/Fiber (Specify your framework)

PDF Processing using pdfplumber, PyMuPDF, or another library

