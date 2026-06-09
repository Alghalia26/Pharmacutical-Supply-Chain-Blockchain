'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class TransferBatchWorkload extends WorkloadModuleBase {
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        this.batchIDs = [];
        this.counter = 0;

        // Pre-create batches before the test starts
        for (let i = 0; i < 200; i++) {
            const batchID = 'TRANSFER_ONLY_' + this.workerIndex + '_' + Date.now() + '_' + i;
            this.batchIDs.push(batchID);

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
                    '100',
                    'DHL Oman'
                ]
            });
        }
    }

    async submitTransaction() {
        const batchID = this.batchIDs[this.counter % this.batchIDs.length];
        this.counter++;

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
    }
}

function createWorkloadModule() {
    return new TransferBatchWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;