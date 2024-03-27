
$(document).ready(function () {

  syncTitle();

  syncJson('about', 'api/about');

  syncTable('deployments', 'api/deployment/list', [
    ['id', 'api/deployment?id=', 'deployments'],
    ['template', 'api/template?id=', 'templates'],
    ['parent', 'api/deployment?id=', 'deployments'],
    ['site', 'api/site?id=', 'sites'],
    ['metadata'],
    ['prepared'],
    ['approved'],
    ['createdTimestamp'],
    ['updatedTimestamp']
  ]);

  syncTable('sites', 'api/site/list', [
    ['id', 'api/site?id=', 'sites'],
    ['template', 'api/template?id=', 'templates'],
    ['metadata'],
    ['deployments', 'api/deployment?id=', 'deployments'],
    ['updatedTimestamp']
  ]);

  syncTable('templates', 'api/template/list', [
    ['id', 'api/template?id=', 'templates'],
    ['metadata'],
    ['deployments', 'api/deployment?id=', 'deployments'],
    ['updatedTimestamp']
  ]);

  syncTable('plugins', 'api/plugin/list', [
    ['type'],
    ['name'],
    ['executor'],
    ['arguments'],
    ['properties'],
    ['triggers']
  ]);

  closeButton('deployments');
  closeButton('sites');
  closeButton('templates');

  $('#deployments-tab').tab('show');

});

const INTERVAL = 2000;

var intervals = {};

var locale = Intl.DateTimeFormat().resolvedOptions().locale;

var dateTimeFormat = new Intl.DateTimeFormat(locale, {dateStyle: 'short', timeStyle: 'short'});

function syncTitle() {
  const description = $('#description');

  function tick() {
    $.get({
      url: 'api/about',
      dataType: 'json',
      success: function (content) {
        let text = '';
        if (content.instanceName)
          text = escapeContent(content.instanceName);
        if (content.instanceDescription)
          text += '<br/>' + escapeContent(content.instanceDescription);
        description.html(text);
      }
    });
  }

  tick();
  intervals['description'] = setInterval(tick, INTERVAL);
}

function syncJson(tab, url) {
  const tabControl = $('#'+tab+'-tab');

  tabControl.on('show.bs.tab', function () {
    showJsonTab(tab, url);
  });

  tabControl.on('hide.bs.tab', function () {
    hideTab(tab);
  });
}

function syncTable(tab, url, columns) {
  const tabControl = $('#'+tab+'-tab');

  tabControl.on('show.bs.tab', function () {
    showTableTab(tab, url, columns);
  });

  tabControl.on('hide.bs.tab', function () {
    hideTab(tab);
  });
}

function closeButton(tab) {
  $('#'+tab+'-close').click(function () {
    $('#'+tab+'-details').hide();
    $('#'+tab+'-list').show();
  });
}

function showJsonTab(tab, url) {
  const json = $('#'+tab+'-json');

  function tick() {
    $.get({
      url: url,
      dataType: 'json',
      success: function (content) {
        content = highlight(JSON.stringify(content, null, '  '), 'json');
        json.html(content);
      }
    }).fail(function () {
      tbody.empty();
    });
  }

  tick();
  intervals[tab] = setInterval(tick, INTERVAL);
}

function showTableTab(tab, url, columns) {
  const tbody = $('#'+tab+' table tbody');

  function tick() {
    $.get({
      url: url,
      dataType: 'json',
      success: function (rows) {
        updateTable(tbody, rows, columns);
      }
    }).fail(function () {
      tbody.empty();
    });
  }

  tick();
  intervals[tab] = setInterval(tick, INTERVAL);
}

function hideTab(tab) {
  clearInterval(intervals[tab]);
  delete intervals[tab];
}

function updateTable(tbody, rows, columns) {
  const ids = rows.map(row => row.id);

  // Removed rows
  eachTableRow(tbody, function (tr) {
    if (!ids.includes(tr.data('id')))
      tr.remove();
    return true;
  });

  for (const row of rows) {
    let newRow = true;

    // Existing row
    eachTableRow(tbody, function (tr) {
      if (tr.data('id') == row.id) {
        newRow = false;
        // Replace row if necessary
        if (!deepEqual(tr.data('row'), row)) {
          const newTr = createTr(row, columns);
          tr.replaceWith(newTr);
        }
        return false;
      }
      return true;
    });

    // New row
    if (newRow) {
      const newTr = createTr(row, columns);

      eachTableRow(tbody, function (tr) {
        // Insert before higher order row
        if (tr.data('id') > row.id) {
          newRow = false;
          tr.before(newTr);
          return false;
        }
        return true;
      });

      // Or append at bottom
      if (newRow)
        tbody.append(newTr);
    }
  }
}

function eachTableRow(tbody, f) {
  tbody.children('tr').each(function () {
    return f($(this));
  });
}

function createTr(row, columns) {
  const tr = $('<tr></tr>');

  tr.data('id', row.id);
  tr.data('row', row);

  for (const column of columns) {
    const [name, urlPrefix, tab] = column;

    const value = row[name];
    const td = $('<td></td>');

    if (urlPrefix && value)
      // Array of links
      if (Array.isArray(value)) {
        let first = true;
        value.forEach(function (e) {
          if (!first)
            td.append('<br/>');
          td.append(newLink(e, urlPrefix, tab));
          first = false;
        });
      } else
        // Single link
        td.append(newLink(value, urlPrefix, tab));
    else if (name.endsWith('Timestamp'))
      td.append(escapeContent(dateTimeFormat.format(new Date(value))));
    else
      // Plain content
      td.append(renderContent(value));
    tr.append(td);
  }

  return tr;
}

function newLink(value, urlPrefix, tab) {
  const link = $('<span class="entity-link">' + renderContent(value) + '</span>');
  link.click(function () {
    $.get({
      url: urlPrefix + value,
      dataType: 'text',
      success: function(content) {
        // Show details tab
        content = highlight(content, 'yaml');
        $('#'+tab+'-title').html(value);
        $('#'+tab+'-yaml').html(content);
        $('#'+tab+'-list').hide();
        $('#'+tab+'-details').show();
        $('#'+tab+'-tab').tab('show');
      }
    })
  });
  return link;
}

function renderContent(value) {
  if ((value === undefined) || (value === null))
    return '';
  else if (Array.isArray(value))
    return value.map(row => renderContent(row)).join('<br/>');

  switch (typeof value) {
  case 'boolean':
    return '<span class="value-'+value+'">'+value+'</span>'
  case 'object':
    return Object.entries(value).toSorted().map(([k, v]) => escapeContent(k)+'='+escapeContent(v)).join('<br/>');
  default:
    return escapeContent(String(value));
  }
}

function escapeContent(html) {
  return html.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

// Unused for now
function escapeAttribute(html) {
  return escapeContent(html).replace(/"/g, '&quot;');
}

function highlight(value, language) {
  return hljs.highlight(value, {language: language}).value
}

function deepEqual(v1, v2) {
  if (v1 === v2) // for non-objects, including undefined and null
    return true;

  if ((v1 === undefined) || (v2 === undefined) || (v1 === null) || (v2 === null) || (typeof v1 !== 'object') || (typeof v2 !== 'object'))
    return false;

  let entries1 = Object.entries(v1); // also works on arrays
  if (entries1.length !== Object.entries(v2).length)
    return false;

  for (const [key1, value1] of entries1) {
    const value2 = v2[key1];
    if ((value2 === undefined) || !deepEqual(value1, value2))
      return false;
  }

  return true;
}
