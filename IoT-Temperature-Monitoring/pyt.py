import random
import time
import subprocess

# ask user for batch (GENERAL)
batch_id = input("Enter Batch ID: ")

while True:
    # generate temperature
    temp = round(random.uniform(1, 10), 2)

    # classify
    if temp < 2:
        status = "Too Low"
    elif temp > 8:
        status = "Too High"
    else:
        status = "Normal"

    print(f"Batch: {batch_id} | Temp: {temp}°C | Status: {status}")

    # send to blockchain
    command = f'''
cd ~/Videos/test/BasicNetwork-2.0
source deployChaincode.sh
updateTemperature {batch_id} {temp}
'''

    subprocess.run(["bash", "-lc", command])

    # wait
    time.sleep(5)