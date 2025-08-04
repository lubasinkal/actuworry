
document.addEventListener('DOMContentLoaded', () => {
    // Pricing Form
    const pricingForm = document.getElementById('pricing-form');
    const premiumResult = document.getElementById('premium-result');
    const premiumValue = document.getElementById('premium-value');
    const pricingSubmitButton = pricingForm.querySelector('button[type="submit"]');

    pricingForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        pricingSubmitButton.disabled = true;
        pricingSubmitButton.textContent = 'Calculating...';

        const formData = new FormData(pricingForm);
        const data = Object.fromEntries(formData.entries());

        try {
            const response = await fetch('/pricing/predict', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data),
            });
            const result = await response.json();
            premiumValue.textContent = result.premium.toFixed(2);
            premiumResult.classList.remove('hidden');
        } catch (error) {
            console.error('Error:', error);
        }

        pricingSubmitButton.disabled = false;
        pricingSubmitButton.textContent = 'Calculate Premium';
    });
});
