
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

    // Navigation and button functionality
    const pricingLink = document.querySelector('a[href="#pricing"]');
    const reservingLink = document.querySelector('a[href="#reserving"]');
    const pricingSection = document.getElementById('pricing');
    const reservingSection = document.getElementById('reserving');
    const sensitivityBtn = document.getElementById('sensitivity-btn');
    const comparisonBtn = document.getElementById('comparison-btn');

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
            term: parseInt(formData.get('term')) || 1, // Default to 1 for annuities
            sum_assured: parseFloat(formData.get('sum_assured')),
            interest_rate: parseFloat(formData.get('interest_rate')) / 100, // Convert percentage to decimal
            table_name: formData.get('table_name'),
            product_type: formData.get('product_type')
        };

        // Add optional underwriting fields
        const smokerStatus = formData.get('smoker_status');
        const healthRating = formData.get('health_rating');
        const ratingFactor = formData.get('rating_factor');
        const deferralPeriod = formData.get('deferral_period');

        if (smokerStatus) data.smoker_status = smokerStatus;
        if (healthRating) data.health_rating = healthRating;
        if (ratingFactor) data.rating_factor = parseFloat(ratingFactor);
        if (deferralPeriod) data.deferral_period = parseInt(deferralPeriod);

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
            
            // Display premium results with product-specific handling
            displayPremiumResults(result);
            
            // Display underwriting and risk information
            displayUnderwritingInfo(result);
            
            // Display annuity-specific information
            displayAnnuityInfo(result);
            
            // Display expense breakdown (only for life insurance)
            const expenseSection = document.getElementById('expense-breakdown');
            if (result.expenses) {
                expenseSection.classList.remove('hidden');
                initialExpenseSpan.textContent = (result.expenses.initial_expense_rate * 100).toFixed(1);
                renewalExpenseSpan.textContent = (result.expenses.renewal_expense_rate * 100).toFixed(1);
                maintenanceExpenseSpan.textContent = result.expenses.maintenance_expense.toFixed(0);
                profitMarginSpan.textContent = (result.expenses.profit_margin * 100).toFixed(1);
            } else {
                expenseSection.classList.add('hidden');
            }
            
            // Handle reserve schedule for life insurance products
            const reserveSection = document.getElementById('reserve-schedule');
            if (result.reserve_schedule && result.reserve_schedule.length > 0) {
                reserveSection.classList.remove('hidden');
                const reserveTableBody = document.getElementById('reserve-table-body');
                reserveTableBody.innerHTML = '';
                
                // Create table
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
                
                // Create chart
                createReserveChart(result.reserve_schedule, result.product_type);
            } else {
                reserveSection.classList.add('hidden');
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

// Chart creation function
let reserveChart = null;

function createReserveChart(reserveSchedule, productType) {
    const ctx = document.getElementById('reserveChart').getContext('2d');
    
    // Destroy existing chart if it exists
    if (reserveChart) {
        reserveChart.destroy();
    }
    
    const labels = reserveSchedule.map((_, index) => index);
    const data = reserveSchedule.map(reserve => Math.max(0, reserve)); // Ensure non-negative for better visualization
    
    const chartTitle = productType === 'whole_life' ? 'Whole Life Reserve Schedule' : 'Term Life Reserve Schedule';
    const chartColor = productType === 'whole_life' ? 'rgba(99, 102, 241, 0.8)' : 'rgba(59, 130, 246, 0.8)';
    const borderColor = productType === 'whole_life' ? 'rgba(99, 102, 241, 1)' : 'rgba(59, 130, 246, 1)';
    
    reserveChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Reserve (BWP)',
                data: data,
                backgroundColor: chartColor,
                borderColor: borderColor,
                borderWidth: 2,
                fill: true,
                tension: 0.3
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                title: {
                    display: true,
                    text: chartTitle,
                    font: {
                        size: 16,
                        weight: 'bold'
                    }
                },
                legend: {
                    display: false
                }
            },
            scales: {
                x: {
                    title: {
                        display: true,
                        text: 'Policy Year'
                    }
                },
                y: {
                    title: {
                        display: true,
                        text: 'Reserve Amount (BWP)'
                    },
                    beginAtZero: true
                }
            }
        }
    });
}

// Display premium results based on product type
function displayPremiumResults(result) {
    const netPremiumValue = document.getElementById('net-premium-value');
    const grossPremiumValue = document.getElementById('gross-premium-value');
    const netPremiumContainer = netPremiumValue.closest('.bg-gray-50');
    const grossPremiumContainer = grossPremiumValue.closest('.bg-gray-50');
    
    // Update labels based on product type
    const netLabel = netPremiumContainer.querySelector('h3');
    const grossLabel = grossPremiumContainer.querySelector('h3');
    const netDescription = netPremiumContainer.querySelector('p.text-sm');
    const grossDescription = grossPremiumContainer.querySelector('p.text-sm');
    
    if (result.product_type.includes('annuity')) {
        netLabel.textContent = 'Premium Cost';
        grossLabel.textContent = 'Total Cost';
        netDescription.textContent = 'Single premium for annuity purchase';
        grossDescription.textContent = 'Total cost including fees';
        
        netPremiumValue.textContent = result.total_premium_cost ? result.total_premium_cost.toFixed(2) : result.net_premium.toFixed(2);
        grossPremiumValue.textContent = result.gross_premium.toFixed(2);
        
        // Show annual payout if available
        if (result.annual_payout) {
            const payoutInfo = document.createElement('div');
            payoutInfo.className = 'mt-2 p-2 bg-green-50 rounded';
            payoutInfo.innerHTML = `<strong>Annual Payout:</strong> BWP ${result.annual_payout.toLocaleString()}`;
            grossPremiumContainer.appendChild(payoutInfo);
        }
    } else {
        netLabel.textContent = 'Net Premium';
        grossLabel.textContent = 'Gross Premium';
        netDescription.textContent = 'Pure risk premium without expenses';
        grossDescription.textContent = 'Market premium including expenses & profit';
        
        netPremiumValue.textContent = result.net_premium.toFixed(2);
        grossPremiumValue.textContent = result.gross_premium.toFixed(2);
    }
}

// Display underwriting and risk assessment information
function displayUnderwritingInfo(result) {
    const expenseSection = document.getElementById('expense-breakdown');
    
    // Create or update underwriting section
    let underwritingSection = document.getElementById('underwriting-info');
    if (!underwritingSection) {
        underwritingSection = document.createElement('div');
        underwritingSection.id = 'underwriting-info';
        underwritingSection.className = 'mt-4 bg-purple-50 p-4 rounded-lg';
        expenseSection.parentNode.insertBefore(underwritingSection, expenseSection.nextSibling);
    }
    
    let underwritingHTML = '';
    
    // Show underwriting factors if present
    if (result.underwriting && Object.keys(result.underwriting).length > 0) {
        underwritingHTML += '<h4 class="font-semibold mb-2 text-purple-800">Underwriting Factors:</h4>';
        underwritingHTML += '<div class="grid grid-cols-2 gap-2 text-sm">';
        
        if (result.underwriting.smoker_status) {
            const status = result.underwriting.smoker_status.replace('_', ' ');
            underwritingHTML += `<div>Smoker Status: <span class="font-medium">${status}</span></div>`;
        }
        if (result.underwriting.health_rating) {
            const rating = result.underwriting.health_rating.charAt(0).toUpperCase() + result.underwriting.health_rating.slice(1);
            underwritingHTML += `<div>Health Rating: <span class="font-medium">${rating}</span></div>`;
        }
        if (result.underwriting.custom_rating_factor) {
            underwritingHTML += `<div>Custom Rating: <span class="font-medium">${result.underwriting.custom_rating_factor}x</span></div>`;
        }
        
        underwritingHTML += '</div>';
    }
    
    // Show risk assessment if present
    if (result.risk_assessment) {
        underwritingHTML += '<h4 class="font-semibold mb-2 mt-3 text-purple-800">Risk Assessment:</h4>';
        underwritingHTML += '<div class="grid grid-cols-2 gap-2 text-sm">';
        
        if (result.risk_assessment.risk_multiplier !== 1) {
            const multiplier = result.risk_assessment.risk_multiplier;
            const riskLevel = multiplier > 1 ? 'Higher Risk' : multiplier < 1 ? 'Lower Risk' : 'Standard Risk';
            const color = multiplier > 1 ? 'text-red-600' : multiplier < 1 ? 'text-green-600' : 'text-gray-600';
            underwritingHTML += `<div>Risk Level: <span class="font-medium ${color}">${riskLevel} (${multiplier.toFixed(2)}x)</span></div>`;
        }
        
        if (result.risk_assessment.annual_death_probability) {
            const probability = (result.risk_assessment.annual_death_probability * 100).toFixed(3);
            underwritingHTML += `<div>Annual Mortality: <span class="font-medium">${probability}%</span></div>`;
        }
        
        if (result.risk_assessment.expected_lifetime_years) {
            const lifetime = Math.round(result.risk_assessment.expected_lifetime_years);
            underwritingHTML += `<div>Expected Lifetime: <span class="font-medium">${lifetime} years</span></div>`;
        }
        
        underwritingHTML += '</div>';
    }
    
    if (underwritingHTML) {
        underwritingSection.innerHTML = underwritingHTML;
        underwritingSection.classList.remove('hidden');
    } else {
        underwritingSection.classList.add('hidden');
    }
}

// Display annuity-specific information
function displayAnnuityInfo(result) {
    if (!result.product_type.includes('annuity')) {
        return;
    }
    
    // Create or update annuity section
    let annuitySection = document.getElementById('annuity-info');
    if (!annuitySection) {
        annuitySection = document.createElement('div');
        annuitySection.id = 'annuity-info';
        annuitySection.className = 'mt-4 bg-green-50 p-4 rounded-lg';
        document.getElementById('premium-result').appendChild(annuitySection);
    }
    
    let annuityHTML = '<h4 class="font-semibold mb-3 text-green-800">Annuity Details:</h4>';
    annuityHTML += '<div class="grid grid-cols-1 md:grid-cols-2 gap-4">';
    
    if (result.annual_payout) {
        annuityHTML += `
            <div class="bg-white p-3 rounded shadow-sm">
                <h5 class="font-medium text-green-700">Annual Payout</h5>
                <p class="text-2xl font-bold text-green-600">BWP ${result.annual_payout.toLocaleString()}</p>
                <p class="text-sm text-gray-600">Per year for life</p>
            </div>
        `;
    }
    
    if (result.total_premium_cost) {
        const payoutYears = Math.round(result.total_premium_cost / result.annual_payout);
        annuityHTML += `
            <div class="bg-white p-3 rounded shadow-sm">
                <h5 class="font-medium text-green-700">Breakeven Period</h5>
                <p class="text-2xl font-bold text-green-600">${payoutYears} years</p>
                <p class="text-sm text-gray-600">To recover premium cost</p>
            </div>
        `;
    }
    
    // Show product-specific information
    if (result.product_type === 'immediate_annuity') {
        annuityHTML += `
            <div class="col-span-full bg-blue-50 p-3 rounded">
                <p class="text-sm"><strong>Immediate Annuity:</strong> Payments start immediately after purchase. 
                Ideal for retirees seeking immediate income.</p>
            </div>
        `;
    } else if (result.product_type === 'deferred_annuity') {
        annuityHTML += `
            <div class="col-span-full bg-blue-50 p-3 rounded">
                <p class="text-sm"><strong>Deferred Annuity:</strong> Payments start after the deferral period. 
                Lower cost now, future income stream.</p>
            </div>
        `;
    }
    
    annuityHTML += '</div>';
    annuitySection.innerHTML = annuityHTML;
    annuitySection.classList.remove('hidden');
}
