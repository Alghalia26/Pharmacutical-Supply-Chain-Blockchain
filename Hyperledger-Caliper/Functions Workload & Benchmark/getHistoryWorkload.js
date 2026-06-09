'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class GetHistoryWorkload extends WorkloadModuleBase {
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        this.batchIDs = [];
        this.counter = 0;

        for (let i = 0; i < 30; i++) {
            const batchID = 'HISTORY_' + this.workerIndex + '_' + Date.now() + '_' + i;
            this.batchIDs.push(batchID);

            // Create batch
            await this.sutAdapter.sendRequests({
                contractId: 'pharmachain',
                contractFunction: 'createBatch',
                invokerIdentity: 'client0',
                contractArguments: [
                    batchID,
                    'Panadol',
                    'Germany',
                    'ShippingOrg1',
                    'At Shipping',
                    '2026-01-01',
                    '2027-01-01',
                    '5',
                    '500',
                    'DHL Oman'
                ]
            });

            // Generate transaction history
            await this.sutAdapter.sendRequests({
                contractId: 'pharmachain',
                contractFunction: 'transferBatch',
                invokerIdentity: 'client0',
                contractArguments: [
                    batchID,
                    'DistributorOrg2',
                    'Muscat Medical Distributor'
                ]
            });

            await this.sutAdapter.sendRequests({
                contractId: 'pharmachain',
                contractFunction: 'updateTemperature',
                invokerIdentity: 'client0',
                contractArguments: [
                    batchID,
                    '6'
                ]
            });
        }
    }

    async submitTransaction() {
        const batchID = this.batchIDs[this.counter % this.batchIDs.length];
        this.counter++;

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'getHistoryForBatch',
            invokerIdentity: 'client0',
            contractArguments: [batchID],
            readOnly: true
        });
    }
}

function createWorkloadModule() {
    return new GetHistoryWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
