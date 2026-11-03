// Actuworry frontend — one Alpine component, one source of truth.
const AMBER = '#e8912a';
const MUTE = '#8a8e96';
const DIM = '#5d6168';
const LINE = '#1f2125';
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
      deferral_period: 10,
    },
    premium: null,

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

    init() {
      if (window.Chart) {
        Chart.defaults.font.family = "Geist, system-ui, sans-serif";
        Chart.defaults.color = MUTE;
      }
    },

    setTab(id) {
      this.tab = id;
      this.error = '';
      this.$nextTick(() => {
        if (id === 'benchmarks') this.renderPerfChart();
        Object.values(charts).forEach((c) => c.resize());
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
        });
        this.$nextTick(() => this.renderReserveChart());
      });
    },

    calculatePV() {
      this.error = '';
      const factor = this.discountFactor();
      this.pvResult = this.pv.amount * factor;
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

    renderReserveChart() {
      if (!this.premium || !this.premium.reserve_schedule) return;
      const values = this.premium.reserve_schedule;
      drawChart('reserve', 'reserveChart', {
        type: 'line',
        data: {
          labels: values.map((_, i) => i),
          datasets: [{
            data: values,
            borderColor: AMBER,
            backgroundColor: 'rgba(232, 145, 42, 0.08)',
            fill: true,
            tension: 0.3,
            pointRadius: 0,
            borderWidth: 1.5,
          }],
        },
        options: lineOptions(),
      });
    },

    renderPVChart() {
      const labels = [];
      const data = [];
      const years = Math.max(1, Math.min(this.pv.years, 60));
      for (let i = 0; i <= years; i++) {
        labels.push(i);
        data.push(this.pv.amount * Math.pow(1 + this.pv.rate / 100, -i));
      }
      drawChart('pv', 'pvChart', {
        type: 'line',
        data: {
          labels,
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
      });
    },

    renderReserveScheduleChart() {
      const values = this.reserveSchedule || [];
      drawChart('reserveSchedule', 'reserveScheduleChart', {
        type: 'line',
        data: {
          labels: values.map((_, i) => i),
          datasets: [{
            data: values,
            borderColor: AMBER,
            backgroundColor: 'rgba(232, 145, 42, 0.08)',
            fill: true,
            tension: 0.3,
            pointRadius: 0,
            borderWidth: 1.5,
          }],
        },
        options: lineOptions(),
      });
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

    renderPerfChart() {
      drawChart('perf', 'perfChart', {
        type: 'bar',
        data: {
          labels: ['Go + v-star', 'Python', 'R', 'Excel VBA', 'Excel'],
          datasets: [{
            data: [28000000, 500000, 100000, 50000, 1000000],
            backgroundColor: [AMBER, '#3a3e45', '#3a3e45', '#3a3e45', '#3a3e45'],
            borderRadius: 4,
            maxBarThickness: 28,
          }],
        },
        options: {
          indexAxis: 'y',
          responsive: true,
          maintainAspectRatio: false,
          animation: false,
          plugins: { legend: { display: false } },
          scales: {
            x: {
              type: 'logarithmic',
              grid: { color: LINE },
              border: { display: false },
              ticks: {
                color: MUTE,
                font: { size: 11 },
                callback: (value) => {
                  if (value >= 1000000) return value / 1000000 + 'M';
                  if (value >= 1000) return value / 1000 + 'K';
                  return value;
                },
              },
            },
            y: {
              grid: { display: false },
              border: { display: false },
              ticks: { color: MUTE, font: { size: 11 } },
            },
          },
        },
      });
    },
  };
}

function drawChart(key, canvasId, config) {
  const canvas = document.getElementById(canvasId);
  if (!canvas || !window.Chart) return;
  if (charts[key]) charts[key].destroy();
  charts[key] = new Chart(canvas, config);
}

function lineOptions() {
  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: false,
    interaction: { intersect: false, mode: 'index' },
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: (context) => 'BWP ' + Number(context.parsed.y).toLocaleString('en-US', {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
          }),
        },
      },
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
    animation: false,
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
