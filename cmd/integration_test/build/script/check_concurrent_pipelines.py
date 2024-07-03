import boto3

client = boto3.client('codepipeline')

def is_pipeline_idle(pipeline_name):
    ps = client.get_pipeline_state(name=pipeline_name)
    print(f"Checking if test stage in {pipeline_name} is enabled.")
    for ss in ps['stageStates']:
        if ss['stageName'].startswith('TestEKSA'):
            if ss['inboundTransitionState']['enabled'] == False:
                print(f"TestEKSA stage is disabled, pipeline {pipeline_name} is idle")
                return True

    print("TestEKSA stage is enabled, checking if any stage is in progress.")
    for ss in ps['stageStates']:
        if ss['latestExecution']['status'] == 'InProgress':
            print(f"{ss['stageName']} is in progress, pipeline is not idle")
            return False

    print(f"No stage is in progress, pipeline {pipeline_name} is idle")
    return True

ps = list(map(lambda p: p['name'], filter(lambda p: p['name'].startswith('aws-eks-anywhere-release'), client.list_pipelines()['pipelines'])))
ps.append('aws-eks-anywhere') # main

if any(map(lambda p: not is_pipeline_idle(p), ps)):
    print("Some test pipelines are not idle")
    exit(1)

print("Check completed, all test piplines are idle")