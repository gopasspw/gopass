#!/usr/bin/env python
import subprocess
import os

import depflow

flow = depflow.Depflow()


os.makedirs('messages', exist_ok=True)


@flow.depends(
    depflow.no_file('{}/bin/protoc-gen-go'.format(os.environ['GOPATH'])))
def protocgengo():
    subprocess.check_call([
        'go',
        'get',
        '-u',
        'github.com/golang/protobuf/protoc-gen-go',
    ])


spaths = []
sources = ('config', 'messages', 'storage', 'types')
for source in sources:
    spaths.append('trezor-common/protob/{}.proto'.format(source))


@flow.depends(
    *[depflow.no_file('{}.pb.go'.format(source)) for source in sources],
    *[depflow.file_hash(spath) for spath in spaths],
)
def proto():
    env = os.environ.copy()
    env['PATH'] += ':{}/bin'.format(os.environ['GOPATH'])
    subprocess.check_call([
        'protoc',
        '--go_out=import_path=messages:messages/.',
        '-I/usr/include',
        '-I./trezor-common/protob',
        *spaths,
    ], env=env)