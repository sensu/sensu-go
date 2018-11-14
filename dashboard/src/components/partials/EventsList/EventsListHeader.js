import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import StatusMenu from "/components/partials/StatusMenu";
import ListHeader from "/components/partials/ListHeader";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import DeleteMenuItem from "/components/partials/ToolbarMenuItems/Delete";
import ExecuteMenuItem from "/components/partials/ToolbarMenuItems/QueueExecution";
import ResolveMenuItem from "/components/partials/ToolbarMenuItems/Resolve";
import Select, { Option } from "/components/partials/ToolbarMenuItems/Select";
import SilenceMenuItem from "/components/partials/ToolbarMenuItems/Silence";
import SubmenuItem from "/components/partials/ToolbarMenuItems/Submenu";
import ToolbarMenu from "/components/partials/ToolbarMenu";
import UnsilenceMenuItem from "/components/partials/ToolbarMenuItems/Unsilence";

class EventsListHeader extends React.Component {
  static propTypes = {
    onClickClearSilences: PropTypes.func.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    onClickResolve: PropTypes.func.isRequired,
    onClickRerun: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    selectedItems: PropTypes.array.isRequired,
    rowCount: PropTypes.number.isRequired,
    namespace: PropTypes.shape({
      checks: PropTypes.object,
      entities: PropTypes.object,
    }),
    onChangeQuery: PropTypes.func.isRequired,
  };

  static defaultProps = {
    namespace: null,
  };

  static fragments = {
    event: gql`
      fragment EventsListHeader_event on Event {
        check {
          isSilenced
        }
      }
    `,
    namespace: gql`
      fragment EventsListHeader_namespace on Namespace {
        checks(limit: 1000) {
          nodes {
            name
          }
        }

        entities(limit: 1000) {
          nodes {
            name
          }
        }
      }
    `,
  };

  requeryEntity = newValue => {
    this.props.onChangeQuery({ filter: `Entity.ID == '${newValue}'` });
  };

  requeryCheck = newValue => {
    this.props.onChangeQuery({ filter: `Check.Name == '${newValue}'` });
  };

  requeryHide = newValue => {
    if (newValue === "passing") {
      this.props.onChangeQuery({ filter: `Check.Status != 0` });
    } else if (newValue === "silenced") {
      this.props.onChangeQuery({ filter: `!IsSilenced` });
    } else {
      throw new TypeError(`unknown value ${newValue}`);
    }
  };

  requeryStatus = newValue => {
    if (Array.isArray(newValue)) {
      if (newValue.length === 1) {
        this.props.onChangeQuery({ filter: `Check.Status == ${newValue}` });
      } else {
        const val = newValue.join(",");
        this.props.onChangeQuery({ filter: `Check.Status IN (${val})` });
      }
    } else if (newValue === "") {
      this.props.onChangeQuery(query => query.delete("filter"));
    } else {
      this.props.onChangeQuery({ filter: newValue });
    }
  };

  updateSort = newValue => {
    this.props.onChangeQuery({ order: newValue });
  };

  renderBulkActions = () => {
    const { selectedItems } = this.props;
    const selectedCount = selectedItems.length;
    const selectedSilenced = selectedItems.filter(ev => ev.check.isSilenced);

    const allSelectedSilenced = selectedSilenced.length === selectedCount;
    const allSelectedUnsilenced = selectedSilenced.length === 0;

    return (
      <ToolbarMenu>
        <ToolbarMenu.Item id="resolve" visible="always">
          <ResolveMenuItem
            description="Resolve selected event(s)."
            onClick={this.props.onClickResolve}
          />
        </ToolbarMenu.Item>

        <ToolbarMenu.Item id="re-run" visible="if-room">
          <ExecuteMenuItem
            title="Re-run Checks"
            titleCondensed="Re-run"
            description="Queue adhoc check executions for selected event(s)."
            onClick={this.props.onClickRerun}
          />
        </ToolbarMenu.Item>

        <ToolbarMenu.Item
          id="silence"
          visible={allSelectedSilenced ? "never" : "if-room"}
        >
          <SilenceMenuItem
            disabled={allSelectedSilenced}
            onClick={this.props.onClickSilence}
          />
        </ToolbarMenu.Item>

        <ToolbarMenu.Item
          id="unsilence"
          visible={allSelectedUnsilenced ? "never" : "if-room"}
        >
          <UnsilenceMenuItem
            disabled={allSelectedUnsilenced}
            onClick={this.props.onClickClearSilences}
          />
        </ToolbarMenu.Item>

        <ToolbarMenu.Item id="delete" visible="never">
          {menu => (
            <ConfirmDelete
              identifier={`${selectedCount} ${
                selectedCount === 1 ? "event" : "events"
              }`}
              onSubmit={() => {
                this.props.onClickDelete();
                menu.close();
              }}
            >
              {confirm => (
                <DeleteMenuItem
                  autoClose={false}
                  title="Delete…"
                  onClick={confirm.open}
                />
              )}
            </ConfirmDelete>
          )}
        </ToolbarMenu.Item>
      </ToolbarMenu>
    );
  };

  renderActions = () => {
    const { namespace: ns } = this.props;
    const entities = ns ? ns.entities.nodes.map(e => e.name) : [];
    const checks = ns ? ns.checks.nodes.map(e => e.name) : [];

    return (
      <ToolbarMenu.Autosizer>
        {({ width }) => (
          <ToolbarMenu width={width}>
            <ToolbarMenu.Item id="hide" visible="if-room">
              <Select title="Hide" onChange={this.requeryHide}>
                <Option value="passing">Passing</Option>
                <Option value="silenced">Silenced</Option>
              </Select>
            </ToolbarMenu.Item>

            <ToolbarMenu.Item id="filter-by-entity" visible="if-room">
              <Select title="Entity" onChange={this.requeryEntity}>
                {entities.map(name => <Option key={name} value={name} />)}
              </Select>
            </ToolbarMenu.Item>

            <ToolbarMenu.Item id="filter-by-check" visible="if-room">
              <Select title="Check" onChange={this.requeryCheck}>
                {checks.map(name => <Option key={name} value={name} />)}
              </Select>
            </ToolbarMenu.Item>

            <ToolbarMenu.Item id="filter-by-status" visible="always">
              <SubmenuItem
                autoClose
                title="Status"
                renderMenu={({ anchorEl, close }) => (
                  <StatusMenu
                    anchorEl={anchorEl}
                    onClose={close}
                    onChange={val => {
                      this.requeryStatus(val);
                      close();
                    }}
                  />
                )}
              />
            </ToolbarMenu.Item>

            <ToolbarMenu.Item id="sort" visible="always">
              <Select title="Sort" onChange={this.updateSort}>
                <Option value="LASTOK">Last OK</Option>
                <Option value="SEVERITY">Severity</Option>
                <Option value="NEWEST">Newest</Option>
                <Option value="OLDEST">Oldest</Option>
              </Select>
            </ToolbarMenu.Item>
          </ToolbarMenu>
        )}
      </ToolbarMenu.Autosizer>
    );
  };

  render() {
    const { selectedItems, rowCount, onClickSelect } = this.props;
    const selectedCount = selectedItems.length;

    return (
      <ListHeader
        sticky
        selectedCount={selectedCount}
        rowCount={rowCount}
        onClickSelect={onClickSelect}
        renderBulkActions={this.renderBulkActions}
        renderActions={this.renderActions}
      />
    );
  }
}

export default EventsListHeader;
