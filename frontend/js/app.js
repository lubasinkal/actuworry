// Alpine.js Application for Actuworry
function actuarialApp() {
    return {
        // State
        activeTab: 'calculator',
        mobileMenuOpen: false,  // Add mobile menu state
        loading: false,
        result: null,
        portfolioResult: null,
        sensitivityResult: null,
        portfolio: [],
        charts: {},
        
        // Form data
        policy: {
            age: 35,
            table_name: 'male',
            product_type: 'term_life',
            sum_assured: 100000,
            term: 10,
            interest_rate: 5,
            smoker_status: '',
            health_rating: '',
            deferral_period: 10
        },
        
        sensitivityPolicy: {
            age: 35,
            table_name: 'male',
            product_type: 'term_life',
            sum_assured: 100000,
            term: 10,
            interest_rate: 5
        },
        
        sensitivityParams: {
            interest_rates: '3,4,5,6,7',
            ages: '30,35,40,45,50',
            coverage_amounts: '50000,100000,150000,200000'
        },
        
        // API base URL - update to match new structure
        apiUrl: '/api',
        
        // Initialize
        init() {
            console.log('Actuworry App Initialized');
        },
        
        // Format numbers with commas
        formatNumber(num) {
            if (!num) return '0';
            return parseFloat(num).toLocaleString('en-US', { 
                minimumFractionDigits: 2, 
                maximumFractionDigits: 2 
            });
        },
        
        // Calculate single premium
        async calculatePremium() {
            this.loading = true;
            this.result = null;
            
            try {
                // Convert interest rate from percentage to decimal
                const policyData = {
                    ...this.policy,
                    interest_rate: this.policy.interest_rate / 100
                };
                
                const response = await fetch(`${this.apiUrl}/calculate`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(policyData)
                });
                
                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.error || 'Calculation failed');
                }
                
                this.result = await response.json();
                
                // Update charts after getting results
                this.$nextTick(() => {
                    this.updateCharts();
                });
                
            } catch (error) {
                console.error('Error calculating premium:', error);
                alert('Error: ' + error.message);
            } finally {
                this.loading = false;
            }
        },
        
        // Reset form
        resetForm() {
            this.policy = {
                age: 35,
                table_name: 'male',
                product_type: 'term_life',
                sum_assured: 100000,
                term: 10,
                interest_rate: 5,
                smoker_status: '',
                health_rating: '',
                deferral_period: 10
            };
            this.result = null;
            this.destroyCharts();
        },
        
        // Add policy to portfolio
        addPolicyToPortfolio() {
            this.portfolio.push({
                ...this.policy,
                interest_rate: this.policy.interest_rate / 100
            });
        },
        
        // Analyze portfolio
        async analyzePortfolio() {
            if (this.portfolio.length === 0) return;
            
            this.loading = true;
            this.portfolioResult = null;
            
            try {
                const response = await fetch(`${this.apiUrl}/analyze/portfolio`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ policies: this.portfolio })
                });
                
                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.error || 'Analysis failed');
                }
                
                this.portfolioResult = await response.json();
                
                // Update portfolio charts
                this.$nextTick(() => {
                    this.updatePortfolioCharts();
                });
                
            } catch (error) {
                console.error('Error analyzing portfolio:', error);
                alert('Error: ' + error.message);
            } finally {
                this.loading = false;
            }
        },
        
        // Run sensitivity analysis
        async runSensitivityAnalysis() {
            this.loading = true;
            this.sensitivityResult = null;
            
            try {
                // Parse comma-separated values
                const interestRates = this.sensitivityParams.interest_rates
                    .split(',').map(r => parseFloat(r) / 100);
                const ages = this.sensitivityParams.ages
                    .split(',').map(a => parseInt(a));
                const coverageAmounts = this.sensitivityParams.coverage_amounts
                    .split(',').map(c => parseFloat(c));
                
                const requestData = {
                    base_policy: {
                        ...this.sensitivityPolicy,
                        interest_rate: this.sensitivityPolicy.interest_rate / 100
                    },
                    interest_rates: interestRates,
                    ages: ages,
                    coverage_amounts: coverageAmounts
                };
                
                const response = await fetch(`${this.apiUrl}/calculate/sensitivity`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(requestData)
                });
                
                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.error || 'Sensitivity analysis failed');
                }
                
                this.sensitivityResult = await response.json();
                
                // Update sensitivity charts
                this.$nextTick(() => {
                    this.updateSensitivityCharts();
                });
                
            } catch (error) {
                console.error('Error running sensitivity analysis:', error);
                alert('Error: ' + error.message);
            } finally {
                this.loading = false;
            }
        },
        
        // Update main charts
        updateCharts() {
            if (!this.result) return;
            
            // Reserve Schedule Chart
            const reserveCtx = document.getElementById('reserveChart');
            if (reserveCtx) {
                this.destroyChart('reserve');
                
                const years = Array.from(
                    { length: this.result.reserve_schedule.length }, 
                    (_, i) => i
                );
                
                this.charts.reserve = new Chart(reserveCtx, {
                    type: 'line',
                    data: {
                        labels: years,
                        datasets: [{
                            label: 'Reserve Amount (BWP)',
                            data: this.result.reserve_schedule,
                            borderColor: '#4F46E5',
                            backgroundColor: 'rgba(79, 70, 229, 0.1)',
                            tension: 0.4,
                            fill: true
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: { display: true },
                            tooltip: {
                                callbacks: {
                                    label: (context) => {
                                        return `BWP ${this.formatNumber(context.parsed.y)}`;
                                    }
                                }
                            }
                        },
                        scales: {
                            y: {
                                beginAtZero: true,
                                ticks: {
                                    callback: (value) => `BWP ${value.toLocaleString()}`
                                }
                            },
                            x: {
                                title: {
                                    display: true,
                                    text: 'Policy Year'
                                }
                            }
                        }
                    }
                });
            }
            
            // Premium Breakdown Chart
            const premiumCtx = document.getElementById('premiumChart');
            if (premiumCtx && this.result.expenses) {
                this.destroyChart('premium');
                
                const netPremium = this.result.net_premium;
                const expenses = this.result.gross_premium - this.result.net_premium;
                const profit = expenses * this.result.expenses.profit_margin;
                const otherExpenses = expenses - profit;
                
                this.charts.premium = new Chart(premiumCtx, {
                    type: 'doughnut',
                    data: {
                        labels: ['Net Premium', 'Expenses', 'Profit Margin'],
                        datasets: [{
                            data: [netPremium, otherExpenses, profit],
                            backgroundColor: [
                                '#4F46E5',
                                '#F59E0B',
                                '#10B981'
                            ],
                            borderWidth: 2,
                            borderColor: '#fff'
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: {
                                position: 'bottom'
                            },
                            tooltip: {
                                callbacks: {
                                    label: (context) => {
                                        const label = context.label || '';
                                        const value = this.formatNumber(context.parsed);
                                        const percentage = ((context.parsed / this.result.gross_premium) * 100).toFixed(1);
                                        return `${label}: BWP ${value} (${percentage}%)`;
                                    }
                                }
                            }
                        }
                    }
                });
            }
        },
        
        // Update portfolio charts
        updatePortfolioCharts() {
            if (!this.portfolioResult) return;
            
            // Product Distribution Chart
            const productCtx = document.getElementById('productChart');
            if (productCtx && this.portfolioResult.product_distribution) {
                this.destroyChart('product');
                
                const labels = Object.keys(this.portfolioResult.product_distribution);
                const data = Object.values(this.portfolioResult.product_distribution);
                
                this.charts.product = new Chart(productCtx, {
                    type: 'bar',
                    data: {
                        labels: labels.map(l => l.replace('_', ' ').toUpperCase()),
                        datasets: [{
                            label: 'Number of Policies',
                            data: data,
                            backgroundColor: '#4F46E5'
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: { display: false }
                        },
                        scales: {
                            y: { beginAtZero: true }
                        }
                    }
                });
            }
            
            // Risk Distribution Chart
            const riskCtx = document.getElementById('riskChart');
            if (riskCtx && this.portfolioResult.risk_distribution) {
                this.destroyChart('risk');
                
                const labels = Object.keys(this.portfolioResult.risk_distribution);
                const data = Object.values(this.portfolioResult.risk_distribution);
                
                this.charts.risk = new Chart(riskCtx, {
                    type: 'pie',
                    data: {
                        labels: labels.map(l => l.replace('_', ' ').toUpperCase()),
                        datasets: [{
                            data: data,
                            backgroundColor: [
                                '#10B981',
                                '#F59E0B',
                                '#EF4444'
                            ]
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: {
                                position: 'bottom'
                            }
                        }
                    }
                });
            }
        },
        
        // Update sensitivity charts
        updateSensitivityCharts() {
            if (!this.sensitivityResult || !this.sensitivityResult.analysis) return;
            
            // Interest Rate Sensitivity
            const interestCtx = document.getElementById('interestSensitivityChart');
            if (interestCtx && this.sensitivityResult.analysis.interest_rate) {
                this.destroyChart('interestSensitivity');
                
                const data = this.sensitivityResult.analysis.interest_rate;
                const labels = data.map(d => (d.value * 100).toFixed(1) + '%');
                const netPremiums = data.map(d => d.result.net_premium);
                const grossPremiums = data.map(d => d.result.gross_premium);
                
                this.charts.interestSensitivity = new Chart(interestCtx, {
                    type: 'line',
                    data: {
                        labels: labels,
                        datasets: [{
                            label: 'Net Premium',
                            data: netPremiums,
                            borderColor: '#4F46E5',
                            backgroundColor: 'rgba(79, 70, 229, 0.1)',
                            tension: 0.4
                        }, {
                            label: 'Gross Premium',
                            data: grossPremiums,
                            borderColor: '#7C3AED',
                            backgroundColor: 'rgba(124, 58, 237, 0.1)',
                            tension: 0.4
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: { position: 'top' }
                        },
                        scales: {
                            y: {
                                beginAtZero: false,
                                ticks: {
                                    callback: (value) => `BWP ${value.toLocaleString()}`
                                }
                            }
                        }
                    }
                });
            }
            
            // Age Sensitivity
            const ageCtx = document.getElementById('ageSensitivityChart');
            if (ageCtx && this.sensitivityResult.analysis.age) {
                this.destroyChart('ageSensitivity');
                
                const data = this.sensitivityResult.analysis.age;
                const labels = data.map(d => d.value);
                const netPremiums = data.map(d => d.result.net_premium);
                const grossPremiums = data.map(d => d.result.gross_premium);
                
                this.charts.ageSensitivity = new Chart(ageCtx, {
                    type: 'line',
                    data: {
                        labels: labels,
                        datasets: [{
                            label: 'Net Premium',
                            data: netPremiums,
                            borderColor: '#10B981',
                            backgroundColor: 'rgba(16, 185, 129, 0.1)',
                            tension: 0.4
                        }, {
                            label: 'Gross Premium',
                            data: grossPremiums,
                            borderColor: '#F59E0B',
                            backgroundColor: 'rgba(245, 158, 11, 0.1)',
                            tension: 0.4
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: { position: 'top' }
                        },
                        scales: {
                            y: {
                                beginAtZero: false,
                                ticks: {
                                    callback: (value) => `BWP ${value.toLocaleString()}`
                                }
                            },
                            x: {
                                title: {
                                    display: true,
                                    text: 'Age (years)'
                                }
                            }
                        }
                    }
                });
            }
        },
        
        // Destroy specific chart
        destroyChart(name) {
            if (this.charts[name]) {
                this.charts[name].destroy();
                delete this.charts[name];
            }
        },
        
        // Destroy all charts
        destroyCharts() {
            Object.keys(this.charts).forEach(name => {
                this.destroyChart(name);
            });
        }
    };
}
