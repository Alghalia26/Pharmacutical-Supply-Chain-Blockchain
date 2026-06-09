'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class QueryBatchWorkload extends WorkloadModuleBase {

    async initializeWorkloadModule(workerIndex, totalWorkers,
    roundIndex, roundArguments, sutAdapter, sutContext) {

        await super.initializeWorkloadModule(
            workerIndex,
            totalWorkers,
            roundIndex,
            roundArguments,
            sutAdapter,
            sutContext
        );

        this.batchID =
        'QUERY_' +
        this.workerIndex +
        '_' +
        Date.now();

        // Create batch first
        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',

            contractFunction: 'createBatch',

            invokerIdentity: 'client0',

            contractArguments: [
                this.batchID,
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

    async submitTransaction() {

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',

            contractFunction: 'queryBatch',

            invokerIdentity: 'client0',

            contractArguments: [this.batchID],

            readOnly: true
        });
    }
}

function createWorkloadModule() {
    return new QueryBatchWorkload();
}

module.exports.createWorkloadModule =
createWorkloadModule;
