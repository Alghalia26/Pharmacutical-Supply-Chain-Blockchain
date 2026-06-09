'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class CreateBatchWorkload extends WorkloadModuleBase {
    async submitTransaction() {
        const batchID = 'CREATE_' + this.workerIndex + '_' + Date.now() + '_' + Math.floor(Math.random() * 1000000);

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

function createWorkloadModule() {
    return new CreateBatchWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
