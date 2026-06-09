'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class UpdateTemperatureWorkload extends WorkloadModuleBase {
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        this.batchIDs = [];
        this.counter = 0;

        for (let i = 0; i < 100; i++) {
            const batchID = 'TEMP_ONLY_' + this.workerIndex + '_' + Date.now() + '_' + i;
            this.batchIDs.push(batchID);

            await this.sutAdapter.sendRequests({
                contractId: 'pharmachain',
                contractFunction: 'createBatch',
                invokerIdentity: 'client0',
                contractArguments: [
                    batchID,
                    'Insulin',
                    'Germany',
                    'ShippingOrg1',
                    'At Shipping',
                    '2026-01-01',
                    '2027-01-01',
                    '5',
                    '100',
                    'DHL Oman'
                ]
            });
        }
    }

    async submitTransaction() {
        const batchID = this.batchIDs[this.counter % this.batchIDs.length];
        this.counter++;

        const temperature = (2 + Math.random() * 8).toFixed(2);

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'updateTemperature',
            invokerIdentity: 'client0',
            contractArguments: [
                batchID,
                temperature
            ]
        });
    }
}

function createWorkloadModule() {
    return new UpdateTemperatureWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
