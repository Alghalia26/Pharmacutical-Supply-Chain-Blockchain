'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class PharmacyStockWorkload extends WorkloadModuleBase {
    async submitTransaction() {
        const batchID = 'STOCK_' + this.workerIndex + '_' + Date.now() + '_' + Math.floor(Math.random() * 1000000000);

        const drugs = ['Panadol', 'Insulin', 'Amoxicillin', 'Ibuprofen'];
        const distributors = ['Muscat Medical Distributor', 'Oman Pharma Distributor'];

        const drugName = drugs[Math.floor(Math.random() * drugs.length)];
        const distributorName = distributors[Math.floor(Math.random() * distributors.length)];

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'createBatch',
            invokerIdentity: 'client0',
            contractArguments: [
                batchID, drugName, 'Oman', 'ShippingOrg1', 'At Shipping',
                '2026-01-01', '2027-01-01', '5', '100', 'DHL Oman'
            ]
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'transferBatch',
            invokerIdentity: 'client0',
            contractArguments: [batchID, 'DistributorOrg2', distributorName]
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'transferBatch',
            invokerIdentity: 'clientDistributor',
            contractArguments: [batchID, 'PharmacyOrg3', 'N/A']
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'confirmReceipt',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID, '100']
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'setReorderPoint',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID, '50']
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'dispenseMedicine',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID, '20']
        });


        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'dispenseMedicine',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID, '20']
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'dispenseMedicine',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID, '20']
        });


        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'queryBatch',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID],
            readOnly: true
        });
    }
}

function createWorkloadModule() {
    return new PharmacyStockWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
