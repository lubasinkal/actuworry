// Actuworry frontend — one Alpine component, one source of truth.
const AMBER = '#e8912a';
const MUTE = '#8a8e96';
const DIM = '#5d6168';
const LINE = '#1f2125';
const REDUCED = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
const charts = {};

function actuarialApp() {
  return {
    tab: 'premium',
    loading: false,
    error: '',

    policy: {
      age: 35,
      table_name: 'male',
      product_type: 'term_life',
      sum_assured: 100000,
      term: 20,
      interest_rate: 5,
      smoker_status: '',
      health_rating: '',
      rating_factor: 0,
      deferral_period: 10,
    },
    premium: null,

    expenses: {
      initial_expense_rate: 3,
      renewal_expense_rate: 5,
      maintenance_expense: 50,
      profit_margin: 15,
    },
    annuityLoading: 10,

    pv: { amount: 100000, years: 20, rate: 5 },
    pvResult: null,
    annuity: { age: 50, amount: 12000, deferral: 10, rate: 5 },
    annuityResult: null,

    reserve: { age: 35, table_name: 'male', sum_assured: 100000, term: 20, rate: 5, year: 10 },
    reserveResult: null,
    reserveSchedule: null,

    mc: { num_paths: 10000, drift: 0.02, vol: 0.15, seed: 0 },
    mcResult: null,

    risk: { input: '100, 250, 180, 420, 95, 310, 200, 150, 280, 175, 320, 90, 450', losses: [] },
    riskResult: null,

    rates: { nominal: 5, compounding: 12 },
    rateResult: null,

    dataOpen: false,
    dataLoading: false,
    dataError: '',
    dataTables: [],

    api: {
      selectedId: 'health',
      loading: false,
      touched: false,
      requestId: 0,
      response: '',
      status: '',
      ok: false,
      error: '',
      copied: false,
    },

    apiEndpoints: [
      { id: 'health', method: 'GET', path: '/api/health', desc: 'Service status' },
      { id: 'tables', method: 'GET', path: '/api/tables', desc: 'Loaded mortality tables' },
      {
        id: 'calculate',
        method: 'POST',
        path: '/api/calculate',
        desc: 'Premium and reserve schedule',
        body: {
          age: 35,
          term: 20,
          sum_assured: 100000,
          interest_rate: 0.05,
          table_name: 'male',
          product_type: 'term_life',
          expenses: { initial_expense_rate: 0.03, renewal_expense_rate: 0.05, maintenance_expense: 50, profit_margin: 0.15 },
          annuity_loading: 0.10,
        },
      },
      {
        id: 'batch',
        method: 'POST',
        path: '/api/calculate/batch',
        desc: 'Many policies at once',
        body: {
          policies: [
            { age: 35, term: 20, sum_assured: 100000, interest_rate: 0.05, table_name: 'male', product_type: 'term_life' },
            { age: 42, term: 15, sum_assured: 250000, interest_rate: 0.05, table_name: 'female', product_type: 'whole_life' },
          ],
        },
      },
      {
        id: 'sensitivity',
        method: 'POST',
        path: '/api/calculate/sensitivity',
        desc: 'One policy, varied inputs',
        body: {
          base_policy: { age: 35, term: 20, sum_assured: 100000, interest_rate: 0.05, table_name: 'male', product_type: 'term_life' },
          interest_rates: [0.03, 0.05, 0.07],
          ages: [30, 40, 50],
          coverage_amounts: [50000, 100000, 150000],
        },
      },
      {
        id: 'portfolio',
        method: 'POST',
        path: '/api/analyze/portfolio',
        desc: 'Portfolio summary',
        body: {
          policies: [
            { age: 35, term: 20, sum_assured: 100000, interest_rate: 0.05, table_name: 'male', product_type: 'term_life' },
            { age: 42, term: 15, sum_assured: 250000, interest_rate: 0.05, table_name: 'female', product_type: 'whole_life' },
          ],
        },
      },
      {
        id: 'montecarlo',
        method: 'POST',
        path: '/api/vstar/montecarlo',
        desc: 'Simulated loss distribution',
        body: { num_paths: 1000, drift: 0.02, volatility: 0.15, seed: 42 },
      },
      {
        id: 'risk',
        method: 'POST',
        path: '/api/vstar/risk',
        desc: 'VaR and CTE',
        body: { losses: [100, 250, 180, 420, 95, 310, 200, 150, 280, 175, 320, 90, 450] },
      },
      {
        id: 'reserve-retro',
        method: 'POST',
        path: '/api/vstar/reserve-retro',
        desc: 'Reserve at a policy year',
        body: {
          policy: { age: 35, term: 20, sum_assured: 100000, interest_rate: 0.05, table_name: 'male', product_type: 'term_life' },
          year: 10,
        },
      },
      {
        id: 'rate-convert',
        method: 'POST',
        path: '/api/vstar/rate-convert',
        desc: 'Nominal, effective, force',
        body: { nominal_rate: 0.05, compounding: 12 },
      },
      {
        id: 'duration',
        method: 'POST',
        path: '/api/vstar/duration',
        desc: 'Duration and convexity',
        body: { cash_flows: [100, 100, 100, 1100], rate: 0.05 },
      },
      {
        id: 'endowment',
        method: 'POST',
        path: '/api/vstar/endowment',
        desc: 'Endowment single premium',
        body: { age: 35, term: 20, sum_assured: 100000, rate: 0.05 },
      },
      {
        id: 'bond',
        method: 'POST',
        path: '/api/vstar/bond',
        desc: 'Bond price and duration',
        body: { face_value: 1000, coupon_rate: 0.06, years: 10, yield_to_maturity: 0.05 },
      },
    ],

    get apiSelected() {
      return this.apiEndpoints.find((endpoint) => endpoint.id === this.api.selectedId) || this.apiEndpoints[0];
    },

    init() {
      if (window.Chart) {
        Chart.defaults.font.family = 'Geist, system-ui, sans-serif';
        Chart.defaults.color = MUTE;
      }
    },

    setTab(id) {
      this.tab = id;
      this.error = '';
      this.$nextTick(() => {
        if (id === 'api' && !this.api.touched) this.apiSend();
        Object.values(charts).forEach((chart) => chart.resize());
      });
    },

    async post(path, body) {
      let response;
      try {
        response = await fetch(path, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body),
        });
      } catch (e) {
        throw new Error('Could not reach the calculation server.');
      }

      let data = null;
      try {
        data = await response.json();
      } catch (e) {
        // non-JSON error body
      }
      if (!response.ok) {
        throw new Error((data && data.error) || 'Request failed (' + response.status + ')');
      }
      return data;
    },

    async run(task) {
      this.loading = true;
      this.error = '';
      try {
        await task();
      } catch (e) {
        this.error = e.message || String(e);
      } finally {
        this.loading = false;
      }
    },

    money(value) {
      if (value === null || value === undefined || isNaN(value)) return '—';
      return 'BWP ' + this.num(value);
    },

    num(value, digits = 2) {
      if (value === null || value === undefined || isNaN(value)) return '—';
      let number = Number(value);
      // Floating-point noise around zero should read as zero.
      if (Math.abs(number) < 0.5 * Math.pow(10, -digits)) number = 0;
      return number.toLocaleString('en-US', {
        minimumFractionDigits: digits,
        maximumFractionDigits: digits,
      });
    },

    pct(value, digits = 2) {
      if (value === null || value === undefined || isNaN(value)) return '—';
      return (Number(value) * 100).toFixed(digits) + '%';
    },

    discountFactor() {
      return Math.pow(1 + this.pv.rate / 100, -this.pv.years);
    },

    riskCount() {
      return this.risk.losses.length;
    },

    calculatePremium() {
      return this.run(async () => {
        this.premium = await this.post('/api/calculate', {
          ...this.policy,
          interest_rate: this.policy.interest_rate / 100,
          annuity_loading: this.annuityLoading / 100,
          expenses: {
            initial_expense_rate: this.expenses.initial_expense_rate / 100,
            renewal_expense_rate: this.expenses.renewal_expense_rate / 100,
            maintenance_expense: this.expenses.maintenance_expense,
            profit_margin: this.expenses.profit_margin / 100,
          },
        });
        this.$nextTick(() => this.renderReserveChart());
      });
    },

    resetExpenses() {
      this.expenses = {
        initial_expense_rate: 3,
        renewal_expense_rate: 5,
        maintenance_expense: 50,
        profit_margin: 15,
      };
    },

    resetRating() {
      this.policy.rating_factor = 0;
    },

    calculatePV() {
      this.error = '';
      this.pvResult = this.pv.amount * this.discountFactor();
      this.$nextTick(() => this.renderPVChart());
    },

    calculateAnnuity() {
      return this.run(async () => {
        this.annuityResult = await this.post('/api/calculate', {
          age: this.annuity.age,
          table_name: this.policy.table_name,
          product_type: 'deferred_annuity',
          sum_assured: this.annuity.amount,
          term: this.annuity.deferral + 20,
          deferral_period: this.annuity.deferral,
          interest_rate: this.annuity.rate / 100,
          annuity_loading: this.annuityLoading / 100,
        });
      });
    },

    calculateReserves() {
      return this.run(async () => {
        this.reserve.year = Math.min(Math.max(1, this.reserve.year || 1), this.reserve.term);
        const policy = {
          age: this.reserve.age,
          table_name: this.reserve.table_name,
          product_type: 'term_life',
          sum_assured: this.reserve.sum_assured,
          term: this.reserve.term,
          interest_rate: this.reserve.rate / 100,
        };
        const premium = await this.post('/api/calculate', policy);
        this.reserveSchedule = premium.reserve_schedule || [];
        this.reserveResult = await this.post('/api/vstar/reserve-retro', {
          policy,
          year: this.reserve.year,
        });
        this.$nextTick(() => this.renderReserveScheduleChart());
      });
    },

    runMonteCarlo() {
      return this.run(async () => {
        this.mcResult = await this.post('/api/vstar/montecarlo', {
          num_paths: this.mc.num_paths,
          drift: this.mc.drift,
          volatility: this.mc.vol,
          seed: this.mc.seed,
        });
        this.$nextTick(() => this.renderMCChart());
      });
    },

    calculateRisk() {
      return this.run(async () => {
        const losses = this.risk.input
          .split(',')
          .map((x) => parseFloat(x.trim()))
          .filter((x) => !isNaN(x));

        if (!losses.length) throw new Error('Enter at least one loss value.');

        this.risk.losses = losses;
        this.riskResult = await this.post('/api/vstar/risk', { losses });
        this.$nextTick(() => this.renderRiskChart());
      });
    },

    convertRate() {
      return this.run(async () => {
        this.rateResult = await this.post('/api/vstar/rate-convert', {
          nominal_rate: this.rates.nominal / 100,
          compounding: Number(this.rates.compounding),
        });
        this.$nextTick(() => this.renderRateChart());
      });
    },

    // Data modal
    openData() {
      this.dataOpen = true;
      document.body.classList.add('no-scroll');
      this.loadTables();
    },

    closeData() {
      this.dataOpen = false;
      document.body.classList.remove('no-scroll');
    },

    async loadTables() {
      if (this.dataLoading || this.dataTables.length) return;
      this.dataLoading = true;
      this.dataError = '';
      try {
        const response = await fetch('/api/tables');
        if (!response.ok) throw new Error('Could not load tables.');
        const data = await response.json();
        this.dataTables = data.tables || [];
      } catch (e) {
        this.dataError = e.message || 'Could not load tables.';
      } finally {
        this.dataLoading = false;
      }
    },

    // API explorer
    selectEndpoint(id) {
      this.api.selectedId = id;
      this.api.requestId += 1;
      this.api.loading = false;
      this.api.response = '';
      this.api.status = '';
      this.api.ok = false;
      this.api.error = '';
      this.api.touched = true;
    },

    apiBody() {
      const endpoint = this.apiSelected;
      return endpoint && endpoint.body ? JSON.stringify(endpoint.body, null, 2) : '';
    },

    async apiSend() {
      if (this.api.loading) return;
      const endpoint = this.apiSelected;
      const requestId = ++this.api.requestId;
      this.api.loading = true;
      this.api.touched = true;
      this.api.error = '';
      this.api.response = '';
      this.api.status = '';

      const started = Date.now();
      try {
        const response = await fetch(endpoint.path, {
          method: endpoint.method,
          headers: endpoint.body ? { 'Content-Type': 'application/json' } : undefined,
          body: endpoint.body ? JSON.stringify(endpoint.body) : undefined,
        });

        const text = await response.text();
        if (requestId !== this.api.requestId) return;

        let pretty = text || '(empty response)';
        try {
          pretty = JSON.stringify(JSON.parse(text), null, 2);
        } catch (e) {
          // keep raw text
        }

        this.api.response = pretty;
        this.api.ok = response.ok;
        this.api.status = response.status + ' ' + response.statusText + ' · ' + (Date.now() - started) + 'ms';
      } catch (e) {
        if (requestId === this.api.requestId) this.api.error = 'Server unreachable';
      } finally {
        if (requestId === this.api.requestId) this.api.loading = false;
      }
    },

    curlText() {
      const endpoint = this.apiSelected;
      const body = endpoint.body
        ? " -H 'Content-Type: application/json' -d '" + JSON.stringify(endpoint.body) + "'"
        : '';
      return 'curl -X ' + endpoint.method + ' ' + window.location.origin + endpoint.path + body;
    },

    async copyCurl() {
      const copied = await this.copyText(this.curlText());
      if (!copied) return;
      this.api.copied = true;
      setTimeout(() => {
        this.api.copied = false;
      }, 1400);
    },

    async copyText(text) {
      try {
        await navigator.clipboard.writeText(text);
        return true;
      } catch (e) {
        const area = document.createElement('textarea');
        area.value = text;
        area.style.position = 'fixed';
        area.style.opacity = '0';
        document.body.appendChild(area);
        area.select();
        let copied = false;
        try {
          copied = document.execCommand('copy');
        } catch (err) {
          copied = false;
        }
        area.remove();
        return copied;
      }
    },

    // Charts
    renderReserveChart() {
      if (!this.premium || !this.premium.reserve_schedule) return;
      drawChart('reserve', 'reserveChart', lineChart(this.premium.reserve_schedule));
    },

    renderPVChart() {
      const labels = [];
      const data = [];
      const years = Math.max(1, Math.min(this.pv.years, 60));
      for (let i = 0; i <= years; i++) {
        labels.push(i);
        data.push(this.pv.amount * Math.pow(1 + this.pv.rate / 100, -i));
      }
      drawChart('pv', 'pvChart', lineChart(data, labels));
    },

    renderReserveScheduleChart() {
      drawChart('reserveSchedule', 'reserveScheduleChart', lineChart(this.reserveSchedule || []));
    },

    renderMCChart() {
      if (!this.mcResult) return;
      const r = this.mcResult;
      drawChart('mc', 'mcChart', {
        type: 'bar',
        data: {
          labels: ['Mean', 'VaR 95', 'VaR 99', 'CTE 95', 'CTE 99', 'Max'],
          datasets: [{
            data: [r.mean, r.var_95, r.var_99, r.cte_95, r.cte_99, r.max],
            backgroundColor: [DIM, AMBER, AMBER, '#c97a24', '#c97a24', DIM],
            borderRadius: 4,
            maxBarThickness: 42,
          }],
        },
        options: barOptions(),
      });
    },

    renderRiskChart() {
      if (!this.riskResult || !this.risk.losses.length) return;
      const sorted = [...this.risk.losses].sort((a, b) => a - b);
      const tailStart = Math.floor(sorted.length * 0.95);
      drawChart('risk', 'riskChart', {
        type: 'bar',
        data: {
          labels: sorted.map((_, i) => i + 1),
          datasets: [{
            data: sorted,
            backgroundColor: sorted.map((_, i) => (i >= tailStart ? AMBER : '#3a3e45')),
            borderRadius: 3,
          }],
        },
        options: barOptions(),
      });
    },

    renderRateChart() {
      const nominal = this.rates.nominal / 100;
      const frequencies = [1, 2, 4, 12, 365];
      const labels = ['Annual', 'Semi-annual', 'Quarterly', 'Monthly', 'Daily', 'Continuous'];
      const data = frequencies.map((m) => (Math.pow(1 + nominal / m, m) - 1) * 100);
      data.push((Math.exp(nominal) - 1) * 100);

      drawChart('rate', 'rateChart', {
        type: 'bar',
        data: {
          labels,
          datasets: [{
            data,
            backgroundColor: data.map((_, i) => (i === data.length - 1 ? AMBER : '#3a3e45')),
            borderRadius: 4,
            maxBarThickness: 42,
          }],
        },
        options: barOptions(true),
      });
    },
  };
}

function lineChart(data, labels) {
  return {
    type: 'line',
    data: {
      labels: labels || data.map((_, i) => i),
      datasets: [{
        data,
        borderColor: AMBER,
        backgroundColor: 'rgba(232, 145, 42, 0.08)',
        fill: true,
        tension: 0.3,
        pointRadius: 0,
        borderWidth: 1.5,
      }],
    },
    options: lineOptions(),
  };
}

function drawChart(key, canvasId, config) {
  const canvas = document.getElementById(canvasId);
  if (!canvas || !window.Chart) return;
  if (charts[key]) charts[key].destroy();
  charts[key] = new Chart(canvas, config);
}

function tooltipMoney(context) {
  return 'BWP ' + Number(context.parsed.y).toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

function lineOptions() {
  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: REDUCED ? false : { duration: 350, easing: 'easeOutQuart' },
    interaction: { intersect: false, mode: 'index' },
    plugins: {
      legend: { display: false },
      tooltip: { callbacks: { label: tooltipMoney } },
    },
    scales: {
      x: {
        grid: { display: false },
        border: { display: false },
        ticks: { color: MUTE, font: { size: 11 }, maxTicksLimit: 10 },
      },
      y: {
        grid: { color: LINE },
        border: { display: false },
        ticks: { color: MUTE, font: { size: 11 } },
      },
    },
  };
}

function barOptions(percent = false) {
  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: REDUCED ? false : { duration: 350, easing: 'easeOutQuart' },
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: (context) => {
            const value = Number(context.parsed.y);
            return percent
              ? value.toFixed(3) + '%'
              : 'BWP ' + value.toLocaleString('en-US', {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                });
          },
        },
      },
    },
    scales: {
      x: {
        grid: { display: false },
        border: { display: false },
        ticks: { color: MUTE, font: { size: 11 } },
      },
      y: {
        grid: { color: LINE },
        border: { display: false },
        ticks: {
          color: MUTE,
          font: { size: 11 },
          callback: percent ? (value) => value + '%' : undefined,
        },
      },
    },
  };
}
