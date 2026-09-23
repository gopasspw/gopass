//go:build linux

package secretservice

// collectionIntrospection is the static introspection XML for
// org.freedesktop.Secret.Collection objects.
const collectionIntrospection = `<node>
  <interface name="org.freedesktop.Secret.Collection">
    <method name="Delete">
      <arg name="prompt" type="o" direction="out"/>
    </method>
    <method name="SearchItems">
      <arg name="attributes" type="a{ss}" direction="in"/>
      <arg name="results" type="ao" direction="out"/>
    </method>
    <method name="CreateItem">
      <arg name="properties" type="a{sv}" direction="in"/>
      <arg name="secret" type="(oayays)" direction="in"/>
      <arg name="replace" type="b" direction="in"/>
      <arg name="item" type="o" direction="out"/>
      <arg name="prompt" type="o" direction="out"/>
    </method>
    <signal name="ItemCreated">
      <arg name="item" type="o"/>
    </signal>
    <signal name="ItemDeleted">
      <arg name="item" type="o"/>
    </signal>
    <signal name="ItemChanged">
      <arg name="item" type="o"/>
    </signal>
    <property name="Items" type="ao" access="read"/>
    <property name="Label" type="s" access="readwrite"/>
    <property name="Locked" type="b" access="read"/>
    <property name="Created" type="t" access="read"/>
    <property name="Modified" type="t" access="read"/>
  </interface>
</node>`

// itemIntrospection is the static introspection XML for
// org.freedesktop.Secret.Item objects.
const itemIntrospection = `<node>
  <interface name="org.freedesktop.Secret.Item">
    <method name="Delete">
      <arg name="prompt" type="o" direction="out"/>
    </method>
    <method name="GetSecret">
      <arg name="session" type="o" direction="in"/>
      <arg name="secret" type="(oayays)" direction="out"/>
    </method>
    <method name="SetSecret">
      <arg name="secret" type="(oayays)" direction="in"/>
    </method>
    <property name="Locked" type="b" access="read"/>
    <property name="Attributes" type="a{ss}" access="readwrite"/>
    <property name="Label" type="s" access="readwrite"/>
    <property name="Created" type="t" access="read"/>
    <property name="Modified" type="t" access="read"/>
  </interface>
</node>`

// sessionIntrospection is the static introspection XML for
// org.freedesktop.Secret.Session objects.
const sessionIntrospection = `<node>
  <interface name="org.freedesktop.Secret.Session">
    <method name="Close"/>
  </interface>
</node>`
