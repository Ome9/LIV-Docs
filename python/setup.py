#!/usr/bin/env python3
"""
LIV Document Format Python SDK
Setup script for installation and distribution
"""

from setuptools import setup, find_packages
import os

# Read the README file for long description
def read_readme():
    readme_path = os.path.join(os.path.dirname(__file__), 'README.md')
    if os.path.exists(readme_path):
        with open(readme_path, 'r', encoding='utf-8') as f:
            return f.read()
    return "LIV Document Format Python SDK for automation and batch processing"

# Read requirements from requirements.txt
def read_requirements():
    req_path = os.path.join(os.path.dirname(__file__), 'requirements.txt')
    if os.path.exists(req_path):
        with open(req_path, 'r', encoding='utf-8') as f:
            return [line.strip() for line in f if line.strip() and not line.startswith('#')]
    return []

setup(
    name="liv-document-format",
    version="0.1.0",
    description="Python SDK for LIV Document Format automation and batch processing",
    long_description=read_readme(),
    long_description_content_type="text/markdown",
    author="LIV Document Format Team",
    author_email="team@liv-format.org",
    url="https://github.com/liv-document-format/liv-python",
    packages=find_packages(where="src"),
    package_dir={"": "src"},
    python_requires=">=3.8",
    install_requires=read_requirements(),
    extras_require={
        'dev': [
            'pytest>=7.0.0',
            'pytest-cov>=4.0.0',
            'black>=22.0.0',
            'flake8>=5.0.0',
            'mypy>=1.0.0',
            'sphinx>=5.0.0',
            'sphinx-rtd-theme>=1.0.0',
        ],
        'async': [
            'aiofiles>=22.0.0',
            'asyncio-subprocess>=0.1.0',
        ],
        'validation': [
            'jsonschema>=4.0.0',
            'cerberus>=1.3.0',
        ]
    },
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Text Processing :: Markup",
        "Topic :: Multimedia :: Graphics",
        "Topic :: Internet :: WWW/HTTP :: Dynamic Content",
    ],
    keywords="liv document format automation batch processing cli",
    project_urls={
        "Bug Reports": "https://github.com/liv-document-format/liv-python/issues",
        "Source": "https://github.com/liv-document-format/liv-python",
        "Documentation": "https://liv-python.readthedocs.io/",
    },
    entry_points={
        'console_scripts': [
            'liv-python=liv.cli:main',
        ],
    },
    include_package_data=True,
    zip_safe=False,
)