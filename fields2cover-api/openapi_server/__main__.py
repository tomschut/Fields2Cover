#!/usr/bin/env python3

import connexion
import os
import sys
from openapi_server import encoder


def main():
    app = connexion.App(__name__, specification_dir='./openapi/')
    app.app.json_encoder = encoder.JSONEncoder
    app.add_api('openapi.yaml',
                arguments={'title': 'Fields2Cover API'},
                pythonic_params=True)

    # Use waitress for production, flask dev server for development
    if os.getenv('FLASK_ENV') == 'development':
        app.run(port=8080)
    else:
        from waitress import serve
        print("Starting Fields2Cover API on http://0.0.0.0:8080", flush=True)
        sys.stdout.flush()
        serve(app, host='0.0.0.0', port=8080, threads=4, _quiet=False)


if __name__ == '__main__':
    main()
