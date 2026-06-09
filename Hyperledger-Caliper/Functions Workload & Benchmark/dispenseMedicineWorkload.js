'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class DispenseMedicineWorkload extends WorkloadModuleBase {
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        this.batchIDs = [];
        this.counter = 0;

        for (let i = 0; i < 50; i++) {
            const batchID = 'DISPENSE_ONLY_' + this.workerIndex + '_' + Date.now() + '_' + i;
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
                    '500',
                    'DHL Oman'
                ]
            });

            await this.sutAdapter.sendRequests({
                contractId: 'pharmachain',
                contractFunction: 'transferBatch',
                invokerIdentity: 'client0',
                contractArguments: [
                    batchID,
                    'PharmacyOrg3',
                    'N/A'
                ]
            });

            await this.sutAdapter.sendRequests({
                contractId: 'pharmachain',
                contractFunction: 'confirmReceipt',
                invokerIdentity: 'clientPharmacy',
                contractArguments: [
                    batchID,
                    '500'
                ]
            });
        }
    }

    async submitTransaction() {
        const batchID = this.batchIDs[this.counter % this.batchIDs.length];
        this.counter++;

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'dispenseMedicine',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [
                batchID,
                '1'
            ]
        });
    }
}

function createWorkloadModule() {
    return new DispenseMedicineWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
