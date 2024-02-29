
$(document).ready(function () {

  syncTable('deployments', 'api/deployment/list', [
    ['id', 'api/deployment?id=', 'deployments'],
    ['template', 'api/template?id=', 'templates'],
    ['parent', 'api/deployment?id=', 'deployments'],
    ['site', 'api/site?id=', 'sites'],
    ['metadata'],
    ['prepared'],
    ['approved'],
    ['created'],
    ['updated']
  ]);

  syncTable('sites', 'api/site/list', [
    ['id', 'api/site?id=', 'sites'],
    ['template', 'api/template?id=', 'templates'],
    ['metadata'],
    ['deployments', 'api/deployment?id=', 'deployments']
  ]);

  syncTable('templates', 'api/template/list', [
    ['id', 'api/template?id=', 'templates'],
    ['metadata'],
    ['deployments', 'api/deployment?id=', 'deployments']
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

function syncTable(tab, url, columns) {
  const tabControl = $('#'+tab+'-tab');

  tabControl.on('show.bs.tab', function () {
    showTab(tab, url, columns);
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

function showTab(tab, url, columns) {
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
  // Removed rows
  const ids = rows.map(row => row.id);
  tbody.children('tr').each(function () {
    const tr = $(this);
    if (!ids.includes(tr.data('id')))
      tr.remove();
  });

  for (const row of rows) {
    let newRow = true;

    // Existing rows
    tbody.children('tr').each(function () {
      const tr = $(this);
      if (tr.data('id') == row.id) {
        newRow = false;
        // Replace row if necessary
        if (!deepEqual(tr.data('row'), row))
          tr.replaceWith(createTr(row, columns));
        return false;
      }
    });

    // New rows
    if (newRow) {
      const newTr = createTr(row, columns);

      tbody.children('tr').each(function () {
        const tr = $(this);
        // Insert before higher order row
        if (tr.data('id') > row.id) {
          newRow = false;
          tr.insertBefore(newTr);
          return false;
        }
      });

      // Or append at bottom
      if (newRow)
        tbody.append(newTr);
    }
  }
}

function createTr(row, columns) {
  const tr = $('<tr></tr');
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
      success: function(yaml) {
        // Show details tab
        yaml = hljs.highlight(yaml, {language: 'yaml'}).value;
        $('#'+tab+'-title').html(value);
        $('#'+tab+'-yaml').html(yaml);
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
