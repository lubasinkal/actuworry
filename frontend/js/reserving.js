document.addEventListener('DOMContentLoaded', () => {
    // Chain Ladder Form
    const chainLadderForm = document.getElementById('chain-ladder-form');
    const chainLadderResult = document.getElementById('chain-ladder-result');
    const clReservesValue = document.getElementById('cl-reserves-value');
    const clSubmitButton = chainLadderForm.querySelector('button[type="submit"]');

    chainLadderForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        clSubmitButton.disabled = true;
        clSubmitButton.textContent = 'Calculating...';

        const formData = new FormData(chainLadderForm);
        const data = {
            claims_data: formData.get('claims_data').split('\n').map(row => row.split(',').map(Number).filter(n => !isNaN(n)))
        };

        try {
            const response = await fetch('/reserving/chain-ladder', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data),
            });
            const result = await response.json();
            clReservesValue.textContent = result.reserves.toFixed(2);
            chainLadderResult.classList.remove('hidden');
        } catch (error) {
            console.error('Error:', error);
        }

        clSubmitButton.disabled = false;
        clSubmitButton.textContent = 'Calculate Reserves';
    });

    // Bornhuetter-Ferguson Form
    const bfForm = document.getElementById('bf-form');
    const bfResult = document.getElementById('bf-result');
    const bfReservesValue = document.getElementById('bf-reserves-value');
    const bfSubmitButton = bfForm.querySelector('button[type="submit"]');

    bfForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        bfSubmitButton.disabled = true;
        bfSubmitButton.textContent = 'Calculating...';

        const formData = new FormData(bfForm);
        const data = {
            claims_data: formData.get('claims_data').split('\n').map(row => row.split(',').map(Number).filter(n => !isNaN(n))),
            expected_loss_ratio: parseFloat(formData.get('expected_loss_ratio')),
            premium: formData.get('premium').split(',').map(Number).filter(n => !isNaN(n)),
        };

        try {
            const response = await fetch('/reserving/bornhuetter-ferguson', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data),
            });
            const result = await response.json();
            bfReservesValue.textContent = result.reserves.toFixed(2);
            bfResult.classList.remove('hidden');
        } catch (error) {
            console.error('Error:', error);
        }

        bfSubmitButton.disabled = false;
        bfSubmitButton.textContent = 'Calculate Reserves';
    });
});
