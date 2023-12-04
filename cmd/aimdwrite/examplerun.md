# Example run

This is one way to run aimdwrite. Using this software requires some experimentation per llm model type. Good advice is to go in small steps forward and use LLM for filling up small gaps and helping out from *blank screen syndrome*



Create file *novel.md*
~~~txt
List of chapters for zombie survival novel

$

~~~

Then start program
~~~bash
./aimdwrite -f novel.md
~~~

Wait and then file is updated to status
~~~
List of chapters for zombie survival novel

* The Beginning
* Introduction to the Characters
* Meeting at the Mall
* Fighting off Zombies in the Suburbs
* Traveling Across the Country
* Surviving a Zombie Apocalypse in the City
* Living on an Island
* Hunting and Gathering for Supplies
* Fighting Off a Giant Horde of Zombies
* Creating a Community in a Post-Apocalyptic World
* Taking Down the Leader of a Group of Survivors
* Escaping from a Secret Government Facility
* Confrontation with an Army of Zombie Soldiers
* Fighting Off the Undead in a Desert Wasteland
* Coming Face to Face with a Cure for the Zombie Virus
* The Final Battle Against the Zombies

~~~

Then add few places to insert new lines and comment out first line

~~~
<!-- List of chapters for zombie survival novel-->

This novel tells about zombie pandemia and what happens to few survivors from finland 

# The Beginning
Zombie pandemia starts

$

# Introduction to the Characters
Novel have following characters
$

# Meeting at the Mall
Charactes decide to meet at mall
$

# Fighting off Zombies in the Suburbs
# Traveling Across the Country
# Surviving a Zombie Apocalypse in the City
# Living on an Island
# Hunting and Gathering for Supplies
# Fighting Off a Giant Horde of Zombies
# Creating a Community in a Post-Apocalyptic World
# Taking Down the Leader of a Group of Survivors
# Escaping from a Secret Government Facility
# Confrontation with an Army of Zombie Soldiers
# Fighting Off the Undead in a Desert Wasteland
# Coming Face to Face with a Cure for the Zombie Virus
# The Final Battle Against the Zombies
~~~

After saving and waiting result is

~~~txt
<!-- List of chapters for zombie survival novel-->

This novel tells about zombie pandemia and what happens to few survivors from finland 

# The Beginning
Zombie pandemia starts


People start to get infected and turn into zombies 

Kari, Jukka and Pekka are trying to survive in the wilderness but they soon realize that they need more people to survive. They go back to their homes and find few others who also want to join them 

They start to travel from place to place, finding new friends and enemies along the way 

Kari starts to have a dream of her dead son who tells her what he wants her to do 

Jukka starts to become paranoid and thinks that someone is following them but it turns out to be nothing 

Pekka becomes ill and they need to find medicine, they get lucky and find a doctor who can treat him 

Kari's dream gets stronger and she knows that her son wants her to go to some place but she doesn't know where 

They eventually find the place and it turns out to be a fortified farmhouse. They meet new people who are already living there and they help each other in surviving 

The group starts to get bigger and they need more supplies, they make plans to raid a nearby store 

Kari finally discovers where her son wants her to go and she leaves the fortified farmhouse 

She meets some more people on her way and together they find a place that is safe from zombies 

The group at the fortified farmhouse starts to become weak because of the lack of supplies 

Kari hears about it and decides to go back and help them 

They all unite and decide to stay in one place permanently, they build a new home where they can live safely 

The story ends with them living happily together, Kari finally knowing what her son wanted from her and everyone finding their own purpose in life

# Introduction to the Characters
Novel have following characters
:
- Mikael - main character who survive with his dog in the forest
- Kati - friend of Mikael and former soldier, who have her own group in the city.
- Jani - teenager who join Kati's group after his parents die.
- Jussi - a bit crazy but skilled mechanic who join Kati's group later.
- Tuomas - friend of Mikael and former bartender, who live in the countryside with his daughter.
- Turo - young thief who help Kati's group sometimes.
 
Story begins with Mikael and his dog living alone in the forest trying to survive. He meets a teenager named Jani who joins him. They both meet Tuomas later, who take them to his farmhouse. Kati's group join them later when they need help. From this point story tell what happens to those characters as they try to survive in the world full of zombies and other dangers. 

This novel have quite a lot of violence and gore, so reader should be aware before reading it. It also have some dark humor and serious moments too. Overall, it's interesting story that keeps reader entertained throughout the book.

# Meeting at the Mall
Charactes decide to meet at mall
 because it has enough of food and other things they need
They meet at mall and decide to stay there for a while. But soon they realize that mall is not safe anymore, zombies are everywhere 

They decide to leave the mall and find a new place to live. They travel trough finland in search of safety. 

On their way they meet other survivors and even make some allies. They encounter dangerous situations and have to fight for their lives. But also they find love and hope that maybe this pandemic will end one day 

But in the end zombies are still everywhere, humans are dying every day and hope is fading away 

They decide to stay together and fight till the very end. Even if it means to die trying 

I enjoyed reading this novel because it was well written and had some unexpected twists. It also showed how people can be in crisis situations and what they are willing to do to survive

# Fighting off Zombies in the Suburbs

# Traveling Across the Country

# Surviving a Zombie Apocalypse in the City

# Living on an Island

# Hunting and Gathering for Supplies

# Fighting Off a Giant Horde of Zombies

# Creating a Community in a Post-Apocalyptic World

# Taking Down the Leader of a Group of Survivors

# Escaping from a Secret Government Facility

# Confrontation with an Army of Zombie Soldiers

# Fighting Off the Undead in a Desert Wasteland

# Coming Face to Face with a Cure for the Zombie Virus

# The Final Battle Against the Zombies
~~~

Now LLM created separate characters on each chapter. So it is good idea to include some common information on upper level of hierarchy. One option is to edit file like this

~~~
<!-- List of chapters for zombie survival novel-->

This novel tells about zombie pandemia and what happens to few survivors from finland 
There are three guys Kari, Jukka and Pekka

# The Beginning
Zombie pandemia starts. People start to get infected and turn into zombies 

Kari, Jukka and Pekka are trying to survive in the wilderness but they soon realize that they need more people to survive. They go back to their homes and find few others who also want to join them 

They start to travel from place to place, finding new friends and enemies along the way 

Kari starts to have a dream of her dead son who tells her what he wants her to do 

Jukka starts to become paranoid and thinks that someone is following them but it turns out to be nothing 

Pekka becomes ill and they need to find medicine, they get lucky and find a doctor who can treat him 

Kari's dream gets stronger and she knows that her son wants her to go to some place but she doesn't know where 

They eventually find the place and it turns out to be a fortified farmhouse. They meet new people who are already living there and they help each other in surviving 

The group starts to get bigger and they need more supplies, they make plans to raid a nearby store 

Kari finally discovers where her son wants her to go and she leaves the fortified farmhouse 

She meets some more people on her way and together they find a place that is safe from zombies 

The group at the fortified farmhouse starts to become weak because of the lack of supplies 

Kari hears about it and decides to go back and help them 

They all unite and decide to stay in one place permanently, they build a new home where they can live safely 

The story ends with them living happily together, Kari finally knowing what her son wanted from her and everyone finding their own purpose in life

# Introduction to the Characters

## Kari

Kari is finnish reservist and 
$

## Pekka
Pekka is student at Helsinki University of Technology. Pekka also
$

## Jukka

Jukka is alcoholic hobo 
$


# Meeting at the Mall
Charactes decide to meet at mall
 because it has enough of food and other things they need
They meet at mall and decide to stay there for a while. But soon they realize that mall is not safe anymore, zombies are everywhere 

They decide to leave the mall and find a new place to live. They travel trough finland in search of safety. 

On their way they meet other survivors and even make some allies. They encounter dangerous situations and have to fight for their lives. But also they find love and hope that maybe this pandemic will end one day 

But in the end zombies are still everywhere, humans are dying every day and hope is fading away 

They decide to stay together and fight till the very end. Even if it means to die trying 

I enjoyed reading this novel because it was well written and had some unexpected twists. It also showed how people can be in crisis situations and what they are willing to do to survive

# Fighting off Zombies in the Suburbs

# Traveling Across the Country

# Surviving a Zombie Apocalypse in the City

# Living on an Island

# Hunting and Gathering for Supplies

# Fighting Off a Giant Horde of Zombies

# Creating a Community in a Post-Apocalyptic World

# Taking Down the Leader of a Group of Survivors

# Escaping from a Secret Government Facility

# Confrontation with an Army of Zombie Soldiers

# Fighting Off the Undead in a Desert Wasteland

# Coming Face to Face with a Cure for the Zombie Virus

# The Final Battle Against the Zombies
~~~

Then 

~~~
<!-- List of chapters for zombie survival novel-->

This novel tells about zombie pandemia and what happens to few survivors from finland 
There are three guys Kari, Jukka and Pekka

# The Beginning
Zombie pandemia starts. People start to get infected and turn into zombies 

Kari, Jukka and Pekka are trying to survive in the wilderness but they soon realize that they need more people to survive. They go back to their homes and find few others who also want to join them 

They start to travel from place to place, finding new friends and enemies along the way 

Kari starts to have a dream of her dead son who tells her what he wants her to do 

Jukka starts to become paranoid and thinks that someone is following them but it turns out to be nothing 

Pekka becomes ill and they need to find medicine, they get lucky and find a doctor who can treat him 

Kari's dream gets stronger and she knows that her son wants her to go to some place but she doesn't know where 

They eventually find the place and it turns out to be a fortified farmhouse. They meet new people who are already living there and they help each other in surviving 

The group starts to get bigger and they need more supplies, they make plans to raid a nearby store 

Kari finally discovers where her son wants her to go and she leaves the fortified farmhouse 

She meets some more people on her way and together they find a place that is safe from zombies 

The group at the fortified farmhouse starts to become weak because of the lack of supplies 

Kari hears about it and decides to go back and help them 

They all unite and decide to stay in one place permanently, they build a new home where they can live safely 

The story ends with them living happily together, Kari finally knowing what her son wanted from her and everyone finding their own purpose in life

# Introduction to the Characters

## Kari

Kari is finnish reservist and 
35 years old
Jukka is a police officer and 40 years old
Pekka is a teacher and 38 years old
They are all from northern finland and have known each other since childhood.


Kari and Jukka live in same small town called Yli-Ii and Pekka lives in next town called Pyhäjärvi.

One day zombie pandemic starts and the world changes. 
Kari,Jukka and Pekka try to survive together but it's not easy. Zombies are everywhere and food is scarce. 
They have to find new place to live and to find new allies. They meet some other people like Lasse, Aki and Sauli.

Lasse is a hunter who lives in forest with his dog, Aki is a mechanic and Sauli is an ex-military guy. 
They all help each other to survive but they also have to face difficult decisions.

Should they stay together or should they split up? Can they trust new people they meet on their way? How long can they keep fighting against the undead? Will they find a cure for zombie virus?

This novel is a mix of horror, action and drama. 
It shows how different people react to apocalyptic situation and what kind of choices they make in order to survive.

## Pekka
Pekka is student at Helsinki University of Technology. Pekka also
 has a sister named Sini who is studying in London.
Jukka is a mechanic and he lives with his wife Elina and their son Valtteri. Kari is unemployed and lives alone. He used to be an alcoholic but quit after his parents died in car accident.

Pekka's friend Timo who also studies at Helsinki University of Technology calls Pekka and tells him that he has found a place where they can hide from the zombies. They all meet at the location which is an old fallout shelter.

After few days in the shelter, Sini arrives with two friends Tommi and Anna. The group decides to go out and get supplies but on their way back to the shelter they are attacked by zombies. Jukka saves them and takes them all back to the shelter.

The group starts running out of food and water so they decide to go to Kari's old house which is close to Helsinki airport. On the way they find a grocery store where they stock up on supplies. They also find a gasoline station that is still operational.

Sini suggests going to her university to see if there are any survivors but on their way back, Kari gets injured and they have to take him to the hospital. While there, they meet another survivor named Janne who is an ex-army man. He tells them that he has found a place where they can go for help and he knows a way through the tunnels under Helsinki.

They travel through the tunnels and find a military base that is still operational. They meet with Colonel Kataja who agrees to help them but tells them that they have to stay in the base until they can secure the area around Helsinki.

After few weeks in the base, Pekka gets an idea on how to stop the zombies and he convinces the colonel to try it out. They use explosives to destroy all the zombies in the area around Helsinki airport which also clears the way for helicopters to come and evacuate them to a safer location.

The story ends with Kari saying that they have all survived but he doesn't know if anyone else made it out of Finland alive.

## Jukka

Jukka is alcoholic hobo 

Kari is ex-military man 
Pekka is computer geek 
Together they try to survive in world full of zombies


# Meeting at the Mall
Charactes decide to meet at mall
 because it has enough of food and other things they need
They meet at mall and decide to stay there for a while. But soon they realize that mall is not safe anymore, zombies are everywhere 

They decide to leave the mall and find a new place to live. They travel trough finland in search of safety. 

On their way they meet other survivors and even make some allies. They encounter dangerous situations and have to fight for their lives. But also they find love and hope that maybe this pandemic will end one day 

But in the end zombies are still everywhere, humans are dying every day and hope is fading away 

They decide to stay together and fight till the very end. Even if it means to die trying 

I enjoyed reading this novel because it was well written and had some unexpected twists. It also showed how people can be in crisis situations and what they are willing to do to survive

# Fighting off Zombies in the Suburbs

# Traveling Across the Country

# Surviving a Zombie Apocalypse in the City

# Living on an Island

# Hunting and Gathering for Supplies

# Fighting Off a Giant Horde of Zombies

# Creating a Community in a Post-Apocalyptic World

# Taking Down the Leader of a Group of Survivors

# Escaping from a Secret Government Facility

# Confrontation with an Army of Zombie Soldiers

# Fighting Off the Undead in a Desert Wasteland

# Coming Face to Face with a Cure for the Zombie Virus

# The Final Battle Against the Zombies
~~~