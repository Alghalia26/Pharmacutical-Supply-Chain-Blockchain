'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class TemperatureWorkload extends WorkloadModuleBase {
    async submitTransaction() {
        const batchID = 'TEMP_' + this.workerIndex + '_' + Date.now() + '_' + Math.floor(Math.random() * 1000000000);

        const drugs = ['Insulin', 'Vaccine', 'Antibiotic', 'EyeDrops'];
        const distributors = ['Muscat Medical Distributor', 'Oman Pharma Distributor'];
        const drugName = drugs[Math.floor(Math.random() * drugs.length)];
        const distributorName = distributors[Math.floor(Math.random() * distributors.length)];

        const tempShipping = (Math.random() * (8 - 2) + 2).toFixed(2);
        const tempDistributor = (Math.random() * (10 - 1) + 1).toFixed(2);
        const tempPharmacy = (Math.random() * (8 - 2) + 2).toFixed(2);

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
            contractFunction: 'updateTemperature',
            invokerIdentity: 'client0',
            contractArguments: [batchID, tempShipping]
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'transferBatch',
            invokerIdentity: 'client0',
            contractArguments: [batchID, 'DistributorOrg2', distributorName]
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'updateTemperature',
            invokerIdentity: 'clientDistributor',
            contractArguments: [batchID, tempDistributor]
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
            contractFunction: 'updateTemperature',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID, tempPharmacy]
        });

        await this.sutAdapter.sendRequests({
            contractId: 'pharmachain',
            contractFunction: 'getTemperatureHistory',
            invokerIdentity: 'clientPharmacy',
            contractArguments: [batchID],
            readOnly: true
        });
    }
}

function createWorkloadModule() {
    return new TemperatureWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;


