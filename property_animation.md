# Property Animations

Being able to animate 

- A set of changes over time.
- Script behavior of a property over time.
- RECORD MODE:
    - Record button
    - Stop button: generate animation, restore
    - Edit keyframes
- SNAPSHOT MODE:
    - Snapshot to states of the scene
- PLAYBACK MODE
- Purposes
    - Runtime animation
    - Cutscenes
    - Retarget for different entities


// component definition code

expose_property(COMPONENT_NAME, "color.red", offset_of(light_t, red));

// to set a property

set_property("Light", "color.red", 1.0f);

// called from animation system
// called form a graph
// called from code

light_t light = get_comoponent("Light"):
light.red = 1.0f;


// * FOUNDATION: get_property(), set_property() -- entity system
// * DRIVING PROPERTIES WITH ANIMATIONS
// * RECORD ANIMATIONS

// * APPLY THIS CURVE --- TO THIS ENTITY, COMPONET, PROPERTY

// RECORD MODE
//
// * Scene tells animation system to record the raw data of the component:
// * When it detects a change in The Truth.

recorder->add_watch(entity, component);


// What is the difference between this and an exported animation?