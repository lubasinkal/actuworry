
document.addEventListener('DOMContentLoaded', () => {
    // Pricing Form
    const pricingForm = document.getElementById('pricing-form');
    const premiumResult = document.getElementById('premium-result');
    const netPremiumValue = document.getElementById('net-premium-value');
    const grossPremiumValue = document.getElementById('gross-premium-value');
    const pricingSubmitButton = pricingForm.querySelector('button[type="submit"]');
    
    // Expense breakdown elements
    const initialExpenseSpan = document.getElementById('initial-expense');
    const renewalExpenseSpan = document.getElementById('renewal-expense');
    const maintenanceExpenseSpan = document.getElementById('maintenance-expense');
    const profitMarginSpan = document.getElementById('profit-margin');

    // Navigation functionality
    const pricingLink = document.querySelector('a[href="#pricing"]');
    const reservingLink = document.querySelector('a[href="#reserving"]');
    const pricingSection = document.getElementById('pricing');
    const reservingSection = document.getElementById('reserving');

    pricingLink.addEventListener('click', (e) => {
        e.preventDefault();
        pricingSection.classList.remove('hidden');
        reservingSection.classList.add('hidden');
        pricingLink.classList.add('text-indigo-600');
        reservingLink.classList.remove('text-indigo-600');
    });

    reservingLink.addEventListener('click', (e) => {
        e.preventDefault();
        reservingSection.classList.remove('hidden');
        pricingSection.classList.add('hidden');
        reservingLink.classList.add('text-indigo-600');
        pricingLink.classList.remove('text-indigo-600');
    });

    pricingForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        pricingSubmitButton.disabled = true;
        pricingSubmitButton.textContent = 'Calculating...';

        const formData = new FormData(pricingForm);
        const data = {
            age: parseInt(formData.get('age')),
            term: parseInt(formData.get('term')),
            sum_assured: parseFloat(formData.get('sum_assured')),
            interest_rate: parseFloat(formData.get('interest_rate')) / 100, // Convert percentage to decimal
            table_name: formData.get('table_name')
        };

        try {
            const response = await fetch('/calculate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data),
            });
            
            if (!response.ok) {
                const errorResult = await response.json();
                throw new Error(errorResult.error || 'Calculation failed');
            }
            
            const result = await response.json();
            
            // Display premium results
            netPremiumValue.textContent = result.net_premium.toFixed(2);
            grossPremiumValue.textContent = result.gross_premium.toFixed(2);
            
            // Display expense breakdown
            if (result.expenses) {
                initialExpenseSpan.textContent = (result.expenses.initial_expense_rate * 100).toFixed(1);
                renewalExpenseSpan.textContent = (result.expenses.renewal_expense_rate * 100).toFixed(1);
                maintenanceExpenseSpan.textContent = result.expenses.maintenance_expense.toFixed(0);
                profitMarginSpan.textContent = (result.expenses.profit_margin * 100).toFixed(1);
            }
            
            // Populate reserve schedule table
            const reserveTableBody = document.getElementById('reserve-table-body');
            reserveTableBody.innerHTML = '';
            
            if (result.reserve_schedule) {
                result.reserve_schedule.forEach((reserve, index) => {
                    const row = reserveTableBody.insertRow();
                    const yearCell = row.insertCell(0);
                    const reserveCell = row.insertCell(1);
                    
                    yearCell.textContent = index;
                    yearCell.className = 'py-1';
                    reserveCell.textContent = reserve.toFixed(2);
                    reserveCell.className = 'text-right py-1';
                    
                    if (index % 2 === 1) {
                        row.className = 'bg-white';
                    }
                });
            }
            
            premiumResult.classList.remove('hidden');
            
            // Scroll to results
            premiumResult.scrollIntoView({ behavior: 'smooth' });
            
        } catch (error) {
            console.error('Error:', error);
            alert('Error calculating premium: ' + error.message);
        }

        pricingSubmitButton.disabled = false;
        pricingSubmitButton.textContent = 'Calculate Premiums';
    });
});
